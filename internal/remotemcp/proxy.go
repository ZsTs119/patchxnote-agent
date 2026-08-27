package remotemcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

const (
	codeParseError       = -32700
	codeInvalidRequest   = -32600
	codeInternalError    = -32603
	codeAuthRequired     = -32001
	codeToolError        = -32002
	codePermissionDenied = -32003
	codeRateLimited      = -32004
)

type TokenProvider interface {
	AccessToken(ctx context.Context) (string, bool, error)
	RefreshNow(ctx context.Context) (string, bool, error)
}

type Proxy struct {
	client        *Client
	tokenProvider TokenProvider
	maxLineBytes  int
}

type ProxyOptions struct {
	Client        *Client
	TokenProvider TokenProvider
	MaxLineBytes  int
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcErrorBody   `json:"error,omitempty"`
}

type rpcErrorBody struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

func NewProxy(options ProxyOptions) (*Proxy, error) {
	if options.Client == nil {
		return nil, fmt.Errorf("remote MCP client is required")
	}
	maxLineBytes := options.MaxLineBytes
	if maxLineBytes <= 0 {
		maxLineBytes = int(defaultMaxBodyBytes)
	}
	return &Proxy{
		client:        options.Client,
		tokenProvider: options.TokenProvider,
		maxLineBytes:  maxLineBytes,
	}, nil
}

func (p *Proxy) Serve(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), p.maxLineBytes)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		response, shouldRespond := p.forward(ctx, append([]byte(nil), line...))
		if !shouldRespond {
			continue
		}
		if _, err := stdout.Write(response); err != nil {
			return fmt.Errorf("write remote MCP response: %w", err)
		}
		if len(response) == 0 || response[len(response)-1] != '\n' {
			if _, err := stdout.Write([]byte("\n")); err != nil {
				return fmt.Errorf("write remote MCP response newline: %w", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read remote MCP request: %w", err)
	}
	return nil
}

func (p *Proxy) forward(ctx context.Context, line []byte) ([]byte, bool) {
	requestID, method, isNotification, err := requestEnvelope(line)
	if err != nil {
		return marshalResponse(errorResponse(nil, codeParseError, "parse_error", "invalid JSON-RPC message")), true
	}
	accessToken, hasToken, tokenErr := p.accessToken(ctx)
	if tokenErr != nil {
		if method == "tools/call" {
			return marshalResponse(errorResponse(requestID, codeAuthRequired, "auth_required", "PatchXNote MCP authorization is required")), true
		}
		accessToken = ""
		hasToken = false
	}
	if method == "tools/call" && (!hasToken || accessToken == "") {
		return marshalResponse(errorResponse(requestID, codeAuthRequired, "auth_required", "PatchXNote MCP authorization is required")), true
	}
	response, err := p.client.Do(ctx, line, accessToken)
	if err == nil && response.AuthFailed && p.tokenProvider != nil {
		if refreshed, ok, refreshErr := p.tokenProvider.RefreshNow(ctx); refreshErr == nil && ok && refreshed != "" {
			response, err = p.client.Do(ctx, line, refreshed)
		}
	}
	if err != nil {
		return marshalResponse(errorResponse(requestID, codeToolError, "transport_error", "PatchXNote remote MCP request failed")), true
	}
	if response.NoResponse {
		return nil, false
	}
	if isNotification {
		return nil, false
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return marshalResponse(httpErrorResponse(requestID, response.StatusCode)), true
	}
	if err := validJSONRPCResponse(response.Body, requestID); err != nil {
		return marshalResponse(errorResponse(requestID, codeToolError, "remote_protocol_error", "PatchXNote remote MCP response was invalid")), true
	}
	return response.Body, true
}

func (p *Proxy) accessToken(ctx context.Context) (string, bool, error) {
	if p.tokenProvider == nil {
		return "", false, nil
	}
	return p.tokenProvider.AccessToken(ctx)
}

func requestEnvelope(line []byte) (json.RawMessage, string, bool, error) {
	var request rpcRequest
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		return nil, "", false, err
	}
	if request.JSONRPC != "2.0" || request.Method == "" {
		return request.ID, "", false, fmt.Errorf("invalid JSON-RPC request")
	}
	return responseID(request.ID), request.Method, len(request.ID) == 0, nil
}

func marshalResponse(response rpcResponse) []byte {
	body, err := json.Marshal(response)
	if err != nil {
		return []byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"internal MCP proxy error","data":{"code":"internal_error"}}}`)
	}
	return body
}

func responseID(id json.RawMessage) json.RawMessage {
	if len(id) == 0 {
		return json.RawMessage("null")
	}
	return id
}

func errorResponse(id json.RawMessage, rpcCode int, code string, message string) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      responseID(id),
		Error: &rpcErrorBody{
			Code:    rpcCode,
			Message: message,
			Data: map[string]any{
				"code": code,
			},
		},
	}
}

func httpErrorResponse(id json.RawMessage, statusCode int) rpcResponse {
	switch statusCode {
	case 401:
		return errorResponse(id, codeAuthRequired, "auth_required", "PatchXNote MCP authorization is required")
	case 403:
		return errorResponse(id, codePermissionDenied, "permission_denied", "PatchXNote MCP permission was denied")
	case 429:
		return errorResponse(id, codeRateLimited, "rate_limited", "PatchXNote MCP rate limit exceeded")
	default:
		return errorResponse(id, codeToolError, "remote_http_error", "PatchXNote remote MCP returned an HTTP error")
	}
}
