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

func (c *Client) CreateAgentSetupSession(ctx context.Context, request AgentSetupSessionCreateRequest, idempotencyKey string) (AgentSetupSessionCreated, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return AgentSetupSessionCreated{}, fmt.Errorf("idempotency key is required")
	}
	var response AgentSetupSessionCreated
	err := c.doJSON(ctx, http.MethodPost, "/v1/agent/setup-sessions", nil, "", idempotencyKey, request, &response, false, http.StatusCreated, http.StatusAccepted)
	return response, err
}

func (c *Client) GetAgentSetupSession(ctx context.Context, sessionID string) (AgentSetupSessionStatus, error) {
	if strings.TrimSpace(sessionID) == "" {
		return AgentSetupSessionStatus{}, fmt.Errorf("setup session id is required")
	}
	var response AgentSetupSessionStatus
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/setup-sessions/"+url.PathEscape(sessionID), nil, "", "", nil, &response, false, http.StatusOK)
	return response, err
}

func (c *Client) GetOAuthAuthorizationServer(ctx context.Context) (OAuthAuthorizationServerMetadata, error) {
	var response OAuthAuthorizationServerMetadata
	err := c.doJSON(ctx, http.MethodGet, "/.well-known/oauth-authorization-server", nil, "", "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) ExchangeOAuthCode(ctx context.Context, request OAuthTokenRequest) (OAuthTokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", request.Code)
	values.Set("redirect_uri", request.RedirectURI)
	values.Set("client_id", request.ClientID)
	values.Set("code_verifier", request.CodeVerifier)
	var response OAuthTokenResponse
	err := c.doForm(ctx, "/v1/agent/oauth/token", values, &response, http.StatusOK)
	return response, err
}

func (c *Client) RefreshOAuthToken(ctx context.Context, request OAuthTokenRequest) (OAuthTokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", request.RefreshToken)
	values.Set("client_id", request.ClientID)
	var response OAuthTokenResponse
	err := c.doForm(ctx, "/v1/agent/oauth/token", values, &response, http.StatusOK)
	return response, err
}

func (c *Client) RevokeOAuthToken(ctx context.Context, token string) error {
	values := url.Values{}
	values.Set("token", token)
	return c.doForm(ctx, "/v1/agent/oauth/revoke", values, nil, http.StatusOK)
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

func (c *Client) GetMemoryDeliveryDocument(ctx context.Context, accessToken string, platform string, memoryID string) (AgentDeliveryDocument, error) {
	if strings.TrimSpace(memoryID) == "" {
		return AgentDeliveryDocument{}, fmt.Errorf("memory id is required")
	}
	query, err := optionalPlatformQuery(platform)
	if err != nil {
		return AgentDeliveryDocument{}, err
	}

	var response AgentDeliveryDocument
	err = c.doJSON(ctx, http.MethodGet, "/v1/agent/memories/"+url.PathEscape(memoryID)+"/delivery-document", query, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) GetMemoryModelIO(ctx context.Context, accessToken string, platform string, memoryID string) (AgentModelIOExport, error) {
	if strings.TrimSpace(memoryID) == "" {
		return AgentModelIOExport{}, fmt.Errorf("memory id is required")
	}
	query, err := optionalPlatformQuery(platform)
	if err != nil {
		return AgentModelIOExport{}, err
	}

	var response AgentModelIOExport
	err = c.doJSON(ctx, http.MethodGet, "/v1/agent/memories/"+url.PathEscape(memoryID)+"/model-io", query, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) GetModelRunIOTrace(ctx context.Context, accessToken string, platform string, requestID string) (AgentModelIOExport, error) {
	if strings.TrimSpace(requestID) == "" {
		return AgentModelIOExport{}, fmt.Errorf("request id is required")
	}
	query, err := optionalPlatformQuery(platform)
	if err != nil {
		return AgentModelIOExport{}, err
	}

	var response AgentModelIOExport
	err = c.doJSON(ctx, http.MethodGet, "/v1/agent/model-runs/"+url.PathEscape(requestID)+"/io-trace", query, accessToken, "", nil, &response, true, http.StatusOK)
	return response, err
}

func (c *Client) ListModelIOTraces(ctx context.Context, accessToken string, params ListModelIOTracesParams) (AgentModelIOTracePage, error) {
	if err := validatePlatform(params.Platform); err != nil {
		return AgentModelIOTracePage{}, err
	}
	if params.Limit < 0 || params.Limit > 50 {
		return AgentModelIOTracePage{}, fmt.Errorf("limit must be between 1 and 50 when set")
	}
	query := url.Values{}
	query.Set("platform", params.Platform)
	setOptionalQuery(query, "request_id", params.RequestID)
	setOptionalQuery(query, "task_type", params.TaskType)
	setOptionalQuery(query, "state", params.State)
	setOptionalQuery(query, "recording_id", params.RecordingID)
	setOptionalQuery(query, "event_id", params.EventID)
	setOptionalQuery(query, "business_id", params.BusinessID)
	setOptionalQuery(query, "date_from", params.DateFrom)
	setOptionalQuery(query, "date_to", params.DateTo)
	setOptionalQuery(query, "cursor", params.Cursor)
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}

	var response AgentModelIOTracePage
	err := c.doJSON(ctx, http.MethodGet, "/v1/agent/model-io-traces", query, accessToken, "", nil, &response, true, http.StatusOK)
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

func (c *Client) doForm(
	ctx context.Context,
	endpointPath string,
	values url.Values,
	out any,
	expectedStatus ...int,
) error {
	requestURL := c.resolveURL(endpointPath, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("build api request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send api request: %w", err)
	}
	if !statusIn(resp.StatusCode, expectedStatus) {
		return parseOAuthOrAPIError(resp)
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

func optionalPlatformQuery(platform string) (url.Values, error) {
	platform = strings.TrimSpace(platform)
	if platform == "" {
		return nil, nil
	}
	if err := validatePlatform(platform); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("platform", platform)
	return query, nil
}

func setOptionalQuery(query url.Values, key string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		query.Set(key, value)
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
