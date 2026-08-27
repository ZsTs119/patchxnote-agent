package remotemcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout       = 15 * time.Second
	defaultMaxBodyBytes  = 1 << 20
	protocolVersion      = "2025-06-18"
	defaultServerVersion = "0.0.0-dev"
)

type Client struct {
	endpointURL string
	httpClient  *http.Client
	userAgent   string
	maxBody     int64
}

type Options struct {
	ServerBaseURL string
	HTTPClient    *http.Client
	UserAgent     string
	Timeout       time.Duration
	MaxBodyBytes  int64
}

type Response struct {
	StatusCode int
	Body       []byte
	NoResponse bool
	AuthFailed bool
}

func New(options Options) (*Client, error) {
	baseURL, err := normalizeBaseURL(options.ServerBaseURL)
	if err != nil {
		return nil, err
	}
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		httpClient.Timeout = timeout
	}
	maxBody := options.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = defaultMaxBodyBytes
	}
	if maxBody < 1024 || maxBody > 8<<20 {
		return nil, fmt.Errorf("remote MCP response limit is invalid")
	}
	return &Client{
		endpointURL: strings.TrimRight(baseURL, "/") + "/mcp",
		httpClient:  httpClient,
		userAgent:   strings.TrimSpace(options.UserAgent),
		maxBody:     maxBody,
	}, nil
}

func (c *Client) Do(ctx context.Context, requestBody []byte, accessToken string) (Response, error) {
	if len(bytes.TrimSpace(requestBody)) == 0 {
		return Response{}, fmt.Errorf("remote MCP request body is required")
	}
	if int64(len(requestBody)) > c.maxBody {
		return Response{}, fmt.Errorf("remote MCP request body is too large")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewReader(requestBody))
	if err != nil {
		return Response{}, fmt.Errorf("build remote MCP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("send remote MCP request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNoContent {
		return Response{StatusCode: resp.StatusCode, NoResponse: true}, nil
	}
	contentType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if contentType == "text/event-stream" {
		return Response{}, fmt.Errorf("remote MCP streamable HTTP responses are not supported by this local proxy yet")
	}
	body, err := readLimited(resp.Body, c.maxBody)
	if err != nil {
		return Response{}, err
	}
	if contentType != "" && contentType != "application/json" {
		return Response{}, fmt.Errorf("remote MCP response content type is unsupported")
	}
	response := Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		AuthFailed: resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(bytes.TrimSpace(body)) == 0 {
			return response, fmt.Errorf("remote MCP returned HTTP %d", resp.StatusCode)
		}
	}
	return response, nil
}

func readLimited(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read remote MCP response: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("remote MCP response body is too large")
	}
	return body, nil
}

func normalizeBaseURL(value string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("remote MCP base URL must be an absolute URL without query or fragment")
	}
	if parsed.Scheme == "https" {
		return trimmed, nil
	}
	if parsed.Scheme == "http" && (parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "localhost" || parsed.Hostname() == "::1") {
		return trimmed, nil
	}
	return "", fmt.Errorf("remote MCP base URL must use HTTPS unless it is local loopback")
}

func validJSONRPCResponse(body []byte, requestID json.RawMessage) error {
	var response rpcResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("remote MCP response is not valid JSON-RPC")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("remote MCP response must contain one JSON-RPC object")
	}
	if response.JSONRPC != "2.0" || len(response.ID) == 0 {
		return fmt.Errorf("remote MCP response is invalid")
	}
	if !jsonRawEqual(response.ID, responseID(requestID)) {
		return fmt.Errorf("remote MCP response id mismatch")
	}
	return nil
}

func jsonRawEqual(left json.RawMessage, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(left))
	leftDecoder.UseNumber()
	rightDecoder := json.NewDecoder(bytes.NewReader(right))
	rightDecoder.UseNumber()
	if err := leftDecoder.Decode(&leftValue); err != nil {
		return false
	}
	if err := rightDecoder.Decode(&rightValue); err != nil {
		return false
	}
	leftBody, err := json.Marshal(leftValue)
	if err != nil {
		return false
	}
	rightBody, err := json.Marshal(rightValue)
	if err != nil {
		return false
	}
	return bytes.Equal(leftBody, rightBody)
}
