package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAgentOTPMethods(t *testing.T) {
	var requestAttempts int
	var verifyAttempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/auth/otp/requests":
			requestAttempts++
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if r.Header.Get("Idempotency-Key") != "idem_request_fixture" {
				t.Fatalf("missing idempotency key: %s", r.Header.Get("Idempotency-Key"))
			}
			if r.Header.Get("Authorization") != "" {
				t.Fatalf("otp request must not send authorization")
			}
			writeJSONResponse(t, w, http.StatusAccepted, OTPRequestAccepted{
				RequestID:       "otp_request_fixture",
				Status:          "accepted",
				CooldownSeconds: 60,
			})
		case "/v1/agent/auth/otp/verifications":
			verifyAttempts++
			if r.Header.Get("Idempotency-Key") != "idem_verify_fixture" {
				t.Fatalf("missing verify idempotency key: %s", r.Header.Get("Idempotency-Key"))
			}
			var request AgentOTPVerificationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode verify request: %v", err)
			}
			if request.RequestID != "otp_request_fixture" || request.ClientInstance == "" {
				t.Fatalf("unexpected verify request: %+v", request)
			}
			writeJSONResponse(t, w, http.StatusOK, AgentSessionResponse{
				AccessToken:            strings.Repeat("a", 32),
				AccessExpiresInSeconds: 3600,
				Account: CurrentAccount{
					ID:                   "acct_fixture",
					Status:               "active",
					RegistrationPlatform: "mobile",
					PhoneMasked:          "+86*******0000",
					StateVersion:         2,
				},
				Scopes: []string{"agent:account.read"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 0)
	accepted, err := client.RequestAgentOTP(context.Background(), AgentOTPRequest{
		Phone:          "+86*******0000",
		ClientInstance: "client_instance_fixture",
	}, "idem_request_fixture")
	if err != nil {
		t.Fatalf("request otp: %v", err)
	}
	if accepted.RequestID != "otp_request_fixture" || accepted.CooldownSeconds != 60 {
		t.Fatalf("unexpected otp response: %+v", accepted)
	}

	session, err := client.VerifyAgentOTP(context.Background(), AgentOTPVerificationRequest{
		RequestID:      accepted.RequestID,
		Code:           "000000",
		ClientInstance: "client_instance_fixture",
	}, "idem_verify_fixture")
	if err != nil {
		t.Fatalf("verify otp: %v", err)
	}
	if session.Account.ID != "acct_fixture" || len(session.Scopes) != 1 {
		t.Fatalf("unexpected session response: %+v", session)
	}
	if requestAttempts != 1 || verifyAttempts != 1 {
		t.Fatalf("unexpected attempts request=%d verify=%d", requestAttempts, verifyAttempts)
	}
}

func TestReadProjectionSuccessAndPagination(t *testing.T) {
	credentialMaterial := strings.Repeat("b", 32)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+credentialMaterial {
			t.Fatalf("unexpected authorization header")
		}
		route := strings.TrimPrefix(r.URL.Path, "/patchnote-test-api")
		switch route {
		case "/v1/agent/me":
			writeFixture(t, w, http.StatusOK, "agent_me_success.json")
		case "/v1/agent/recorder-cards":
			writeFixture(t, w, http.StatusOK, "agent_recorder_cards_success.json")
		case "/v1/agent/quota/summary":
			writeFixture(t, w, http.StatusOK, "agent_quota_summary_success.json")
		case "/v1/agent/model-usage/summary":
			writeFixture(t, w, http.StatusOK, "agent_model_usage_summary_success.json")
		case "/v1/agent/memories":
			if r.URL.Query().Get("platform") != "mobile" || r.URL.Query().Get("limit") != "20" || r.URL.Query().Get("cursor") != "cursor_page_1" {
				t.Fatalf("unexpected memories query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_memories_page_success.json")
		case "/v1/agent/memories/mem_fixture_1":
			if r.URL.Query().Get("platform") != "mobile" {
				t.Fatalf("unexpected memory detail query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_memory_success.json")
		case "/v1/agent/auth/logout":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/patchnote-test-api", 0)
	user, err := client.CurrentUser(context.Background(), credentialMaterial)
	if err != nil {
		t.Fatalf("current user: %v", err)
	}
	if user.ID != "acct_fixture" {
		t.Fatalf("unexpected current user: %+v", user)
	}
	cards, err := client.ListRecorderCards(context.Background(), credentialMaterial)
	if err != nil {
		t.Fatalf("recorder cards: %v", err)
	}
	if len(cards.Items) != 1 || cards.Items[0].ID != "hw_fixture" {
		t.Fatalf("unexpected cards: %+v", cards)
	}
	quota, err := client.GetQuotaSummary(context.Background(), credentialMaterial)
	if err != nil {
		t.Fatalf("quota summary: %v", err)
	}
	if quota.AvailableTokens != 120000 {
		t.Fatalf("unexpected quota: %+v", quota)
	}
	usage, err := client.GetModelUsageSummary(context.Background(), credentialMaterial)
	if err != nil {
		t.Fatalf("model usage: %v", err)
	}
	if usage.Period != "current_month" {
		t.Fatalf("unexpected usage: %+v", usage)
	}
	page, err := client.ListMemories(context.Background(), credentialMaterial, ListMemoriesParams{
		Platform: "mobile",
		Limit:    20,
		Cursor:   "cursor_page_1",
	})
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if page.NextCursor != "cursor_page_2" || len(page.Items) != 1 {
		t.Fatalf("unexpected memories page: %+v", page)
	}
	memory, err := client.GetMemory(context.Background(), credentialMaterial, "mobile", "mem_fixture_1")
	if err != nil {
		t.Fatalf("get memory: %v", err)
	}
	if memory.ID != "mem_fixture_1" || memory.SourceAvailability != "metadata_only" {
		t.Fatalf("unexpected memory: %+v", memory)
	}
	if err := client.Logout(context.Background(), credentialMaterial); err != nil {
		t.Fatalf("logout: %v", err)
	}
}

func TestAPIErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		fixture    string
		call       func(*Client) error
		wantCode   string
		wantReqID  string
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			fixture:    "error_unauthorized.json",
			call: func(client *Client) error {
				_, err := client.CurrentUser(context.Background(), strings.Repeat("c", 32))
				return err
			},
			wantCode:  "auth_required",
			wantReqID: "req_fixture_unauthorized",
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			fixture:    "error_forbidden.json",
			call: func(client *Client) error {
				_, err := client.ListRecorderCards(context.Background(), strings.Repeat("c", 32))
				return err
			},
			wantCode:  "permission_denied",
			wantReqID: "req_fixture_forbidden",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			fixture:    "error_not_found.json",
			call: func(client *Client) error {
				_, err := client.GetMemory(context.Background(), strings.Repeat("c", 32), "mobile", "mem_missing")
				return err
			},
			wantCode:  "not_found",
			wantReqID: "req_fixture_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeFixture(t, w, tt.statusCode, tt.fixture)
			}))
			defer server.Close()

			client := newTestClient(t, server.URL, 0)
			err := tt.call(client)
			if err == nil {
				t.Fatal("expected api error")
			}
			var apiErr *Error
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *api.Error, got %T: %v", err, err)
			}
			if apiErr.StatusCode != tt.statusCode || apiErr.Code != tt.wantCode || apiErr.RequestID != tt.wantReqID {
				t.Fatalf("unexpected api error: %+v", apiErr)
			}
			if strings.Contains(err.Error(), apiErr.Message) {
				t.Fatalf("error string should not include server message: %s", err.Error())
			}
		})
	}
}

func TestReadRetryOnlyForRetryableGET(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			writeFixture(t, w, http.StatusTooManyRequests, "error_rate_limited.json")
			return
		}
		writeFixture(t, w, http.StatusOK, "agent_me_success.json")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 1)
	user, err := client.CurrentUser(context.Background(), strings.Repeat("d", 32))
	if err != nil {
		t.Fatalf("current user after retry: %v", err)
	}
	if user.ID != "acct_fixture" {
		t.Fatalf("unexpected user after retry: %+v", user)
	}
	if attempts != 2 {
		t.Fatalf("expected one retry, got %d attempts", attempts)
	}
}

func TestPOSTIsNotRetried(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		writeFixture(t, w, http.StatusTooManyRequests, "error_rate_limited.json")
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 3)
	_, err := client.RequestAgentOTP(context.Background(), AgentOTPRequest{
		Phone:          "+86*******0000",
		ClientInstance: "client_instance_fixture",
	}, "idem_request_fixture")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if attempts != 1 {
		t.Fatalf("expected no retry for POST, got %d attempts", attempts)
	}
}

func TestListMemoriesRequiresPlatform(t *testing.T) {
	client := newTestClient(t, "http://example.test", 0)
	_, err := client.ListMemories(context.Background(), strings.Repeat("e", 32), ListMemoriesParams{})
	if err == nil {
		t.Fatal("expected platform validation error")
	}
}

func newTestClient(t *testing.T, baseURL string, maxReadRetries int) *Client {
	t.Helper()
	client, err := New(Options{
		BaseURL:        baseURL,
		MaxReadRetries: maxReadRetries,
		Sleep: func(ctx context.Context, duration time.Duration) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}

func writeFixture(t *testing.T, w http.ResponseWriter, statusCode int, name string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "api", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
}

func writeJSONResponse(t *testing.T, w http.ResponseWriter, statusCode int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json response: %v", err)
	}
}
