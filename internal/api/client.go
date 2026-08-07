package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

type Client struct {
	baseURL        *url.URL
	httpClient     *http.Client
	userAgent      string
	maxReadRetries int
	sleep          func(context.Context, time.Duration) error
}

type Options struct {
	BaseURL        string
	HTTPClient     *http.Client
	UserAgent      string
	Timeout        time.Duration
	MaxReadRetries int
	Sleep          func(context.Context, time.Duration) error
}

func New(options Options) (*Client, error) {
	if strings.TrimSpace(options.BaseURL) == "" {
		return nil, fmt.Errorf("api base URL is required")
	}
	baseURL, err := url.Parse(options.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse api base URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("api base URL must use http or https")
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf("api base URL host is required")
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

	maxReadRetries := options.MaxReadRetries
	if maxReadRetries < 0 {
		maxReadRetries = 0
	}
	sleep := options.Sleep
	if sleep == nil {
		sleep = sleepContext
	}

	return &Client{
		baseURL:        baseURL,
		httpClient:     httpClient,
		userAgent:      strings.TrimSpace(options.UserAgent),
		maxReadRetries: maxReadRetries,
		sleep:          sleep,
	}, nil
}

func (c *Client) RequestAgentOTP(ctx context.Context, request AgentOTPRequest, idempotencyKey string) (OTPRequestAccepted, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return OTPRequestAccepted{}, fmt.Errorf("idempotency key is required")
	}
	var response OTPRequestAccepted
	err := c.doJSON(ctx, http.MethodPost, "/v1/agent/auth/otp/requests", nil, "", idempotencyKey, request, &response, false, http.StatusAccepted)
	return response, err
}

func (c *Client) VerifyAgentOTP(ctx context.Context, request AgentOTPVerificationRequest, idempotencyKey string) (AgentSessionResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return AgentSessionResponse{}, fmt.Errorf("idempotency key is required")
	}
	var response AgentSessionResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/agent/auth/otp/verifications", nil, "", idempotencyKey, request, &response, false, http.StatusOK)
	return response, err
}

func (c *Client) RefreshAgentSession(ctx context.Context, request AgentRefreshRequest, idempotencyKey string) (AgentSessionResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return AgentSessionResponse{}, fmt.Errorf("idempotency key is required")
	}
	var response AgentSessionResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/agent/auth/refresh", nil, "", idempotencyKey, request, &response, false, http.StatusOK)
	return response, err
}

func (c *Client) CurrentUser(ctx context.Context, accessToken string) (CurrentAccount, error) {
	var response CurrentAccount
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/me", nil, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) ListRecorderCards(ctx context.Context, accessToken string) (AgentRecorderCardPage, error) {
	var response AgentRecorderCardPage
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/recorder-cards", nil, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) GetQuotaSummary(ctx context.Context, accessToken string) (AgentQuotaSummary, error) {
	var response AgentQuotaSummary
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/quota/summary", nil, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) GetModelUsageSummary(ctx context.Context, accessToken string) (AgentModelUsageSummary, error) {
	var response AgentModelUsageSummary
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/model-usage/summary", nil, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) ListMemories(ctx context.Context, accessToken string, params ListMemoriesParams) (AgentMemoryPage, error) {
	if err := validatePlatform(params.Platform); err != nil {
		return AgentMemoryPage{}, err
	}
	query := url.Values{}
	query.Set("platform", params.Platform)
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Cursor != "" {
		query.Set("cursor", params.Cursor)
	}

	var response AgentMemoryPage
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/memories", query, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) GetMemory(ctx context.Context, accessToken string, platform string, memoryID string) (AgentMemory, error) {
	if err := validatePlatform(platform); err != nil {
		return AgentMemory{}, err
	}
	if strings.TrimSpace(memoryID) == "" {
		return AgentMemory{}, fmt.Errorf("memory id is required")
	}
	query := url.Values{}
	query.Set("platform", platform)

	var response AgentMemory
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/memories/"+url.PathEscape(memoryID), query, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) Logout(ctx context.Context, accessToken string) error {
	return c.doJSON(ctx, http.MethodPost, "/v1/agent/auth/logout", nil, accessToken, "", nil, nil, false, http.StatusNoContent)
}

func (c *Client) doJSON(
	ctx context.Context,
	method string,
	endpointPath string,
	query url.Values,
	accessToken string,
	idempotencyKey string,
	body any,
	out any,
	retryableRead bool,
	expectedStatus ...int,
) error {
	var lastErr error
	attempts := 1
	if retryableRead {
		attempts += c.maxReadRetries
	}

	for attempt := 0; attempt < attempts; attempt++ {
		err := c.doJSONOnce(ctx, method, endpointPath, query, accessToken, idempotencyKey, body, out, expectedStatus...)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryableRead || attempt == attempts-1 || !isRetryableAPIError(err) {
			break
		}
		var apiErr *Error
		if errors.As(err, &apiErr) {
			if sleepErr := c.sleep(ctx, apiErr.RetryAfter); sleepErr != nil {
				return sleepErr
			}
		}
	}
	return lastErr
}

func (c *Client) doJSONOnce(
	ctx context.Context,
	method string,
	endpointPath string,
	query url.Values,
	accessToken string,
	idempotencyKey string,
	body any,
	out any,
	expectedStatus ...int,
) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}

	requestURL := c.resolveURL(endpointPath, query)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("build api request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send api request: %w", err)
	}

	if !statusIn(resp.StatusCode, expectedStatus) {
		return parseAPIError(resp)
	}
	defer resp.Body.Close()

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}
	return nil
}

func (c *Client) resolveURL(endpointPath string, query url.Values) string {
	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(path.Clean(endpointPath), "/")
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

func statusIn(statusCode int, expected []int) bool {
	for _, status := range expected {
		if statusCode == status {
			return true
		}
	}
	return false
}

func isRetryableAPIError(err error) bool {
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		return false
	}
	if !apiErr.Retryable {
		return false
	}
	return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode == http.StatusServiceUnavailable
}

func validatePlatform(platform string) error {
	switch platform {
	case "mobile", "desktop":
		return nil
	default:
		return fmt.Errorf("platform must be mobile or desktop")
	}
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
