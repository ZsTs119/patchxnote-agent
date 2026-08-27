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
				AccessToken:             strings.Repeat("a", 32),
				AccessExpiresInSeconds:  3600,
				RefreshToken:            strings.Repeat("r", 43),
				RefreshExpiresInSeconds: 2592000,
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
	if session.Account.ID != "acct_fixture" || len(session.Scopes) != 1 || session.RefreshToken == "" {
		t.Fatalf("unexpected session account=%q scopes=%d has_refresh=%v", session.Account.ID, len(session.Scopes), session.RefreshToken != "")
	}
	if requestAttempts != 1 || verifyAttempts != 1 {
		t.Fatalf("unexpected attempts request=%d verify=%d", requestAttempts, verifyAttempts)
	}
}

func TestRefreshAgentSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agent/auth/refresh" || r.Method != http.MethodPost {
			t.Fatalf("unexpected refresh request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Idempotency-Key") != "idem_refresh_fixture" {
			t.Fatalf("missing refresh idempotency key: %s", r.Header.Get("Idempotency-Key"))
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("refresh request must not send authorization")
		}
		var request AgentRefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode refresh request: %v", err)
		}
		if request.RefreshToken != strings.Repeat("r", 43) {
			t.Fatal("unexpected refresh token body")
		}
		writeJSONResponse(t, w, http.StatusOK, AgentSessionResponse{
			AccessToken:             strings.Repeat("a", 32),
			AccessExpiresInSeconds:  900,
			RefreshToken:            strings.Repeat("n", 43),
			RefreshExpiresInSeconds: 2592000,
			Account: CurrentAccount{
				ID:                   "acct_fixture",
				Status:               "active",
				RegistrationPlatform: "mobile",
				PhoneMasked:          "+86*******0000",
				StateVersion:         2,
			},
			Scopes: []string{"agent:account.read"},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 0)
	session, err := client.RefreshAgentSession(context.Background(), AgentRefreshRequest{RefreshToken: strings.Repeat("r", 43)}, "idem_refresh_fixture")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if session.Account.ID != "acct_fixture" || session.RefreshToken != strings.Repeat("n", 43) {
		t.Fatalf("unexpected refresh account=%q refresh_rotated=%v", session.Account.ID, session.RefreshToken == strings.Repeat("n", 43))
	}
}

func TestOAuthMetadataAndFormMethods(t *testing.T) {
	var metadataCalled bool
	var exchangeCalled bool
	var refreshCalled bool
	var revokeCalled bool
	oldRefresh := strings.Repeat("r", 43)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := strings.TrimPrefix(r.URL.Path, "/patchnote-test-api")
		switch route {
		case "/.well-known/oauth-authorization-server":
			metadataCalled = true
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected metadata method: %s", r.Method)
			}
			if r.Header.Get("Authorization") != "" {
				t.Fatal("metadata request must not send authorization")
			}
			writeJSONResponse(t, w, http.StatusOK, OAuthAuthorizationServerMetadata{
				Issuer:                        serverURLForTest(r),
				AuthorizationEndpoint:         serverURLForTest(r) + "/v1/agent/oauth/authorize",
				TokenEndpoint:                 serverURLForTest(r) + "/v1/agent/oauth/token",
				RevocationEndpoint:            serverURLForTest(r) + "/v1/agent/oauth/revoke",
				ResponseTypesSupported:        []string{"code"},
				GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
				CodeChallengeMethodsSupported: []string{"S256"},
				ScopesSupported:               []string{"agent:account.read"},
			})
		case "/v1/agent/oauth/token":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected token method: %s", r.Method)
			}
			if r.Header.Get("Authorization") != "" {
				t.Fatal("token request must not send authorization")
			}
			if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/x-www-form-urlencoded") {
				t.Fatalf("unexpected content type: %s", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse form: %v", err)
			}
			switch r.PostForm.Get("grant_type") {
			case "authorization_code":
				exchangeCalled = true
				if r.PostForm.Get("code") != strings.Repeat("c", 43) ||
					r.PostForm.Get("redirect_uri") != "http://127.0.0.1:49152/callback" ||
					r.PostForm.Get("client_id") != "patchxnote-local-dev" ||
					r.PostForm.Get("code_verifier") != strings.Repeat("v", 43) {
					t.Fatalf("unexpected exchange form: %s", r.PostForm.Encode())
				}
				if r.PostForm.Get("client_secret") != "" || r.PostForm.Get("access_token") != "" {
					t.Fatalf("exchange form included forbidden field: %s", r.PostForm.Encode())
				}
				writeJSONResponse(t, w, http.StatusOK, OAuthTokenResponse{
					AccessToken:           strings.Repeat("a", 43),
					TokenType:             "Bearer",
					ExpiresIn:             900,
					RefreshToken:          oldRefresh,
					RefreshTokenExpiresIn: 2592000,
					Scope:                 "agent:account.read",
					ConnectorSessionID:    "mcpconn_fixture",
				})
			case "refresh_token":
				refreshCalled = true
				if r.PostForm.Get("refresh_token") != oldRefresh ||
					r.PostForm.Get("client_id") != "patchxnote-local-dev" ||
					r.PostForm.Get("code") != "" || r.PostForm.Get("code_verifier") != "" {
					t.Fatalf("unexpected refresh form: %s", r.PostForm.Encode())
				}
				writeJSONResponse(t, w, http.StatusOK, OAuthTokenResponse{
					AccessToken:           strings.Repeat("b", 43),
					TokenType:             "Bearer",
					ExpiresIn:             900,
					RefreshToken:          strings.Repeat("n", 43),
					RefreshTokenExpiresIn: 2592000,
					Scope:                 "agent:account.read",
					ConnectorSessionID:    "mcpconn_fixture",
				})
			default:
				t.Fatalf("unexpected grant type: %s", r.PostForm.Get("grant_type"))
			}
		case "/v1/agent/oauth/revoke":
			revokeCalled = true
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected revoke method: %s", r.Method)
			}
			if r.Header.Get("Authorization") != "" {
				t.Fatal("revoke request must not send authorization")
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse revoke form: %v", err)
			}
			if r.PostForm.Get("token") != oldRefresh {
				t.Fatal("unexpected revoke token body")
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL+"/patchnote-test-api", 0)
	metadata, err := client.GetOAuthAuthorizationServer(context.Background())
	if err != nil {
		t.Fatalf("oauth metadata: %v", err)
	}
	if metadata.TokenEndpoint == "" || len(metadata.GrantTypesSupported) != 2 {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}

	exchanged, err := client.ExchangeOAuthCode(context.Background(), OAuthTokenRequest{
		Code:         strings.Repeat("c", 43),
		RedirectURI:  "http://127.0.0.1:49152/callback",
		ClientID:     "patchxnote-local-dev",
		CodeVerifier: strings.Repeat("v", 43),
	})
	if err != nil {
		t.Fatalf("exchange oauth code: %v", err)
	}
	if exchanged.TokenType != "Bearer" || exchanged.RefreshToken != oldRefresh {
		t.Fatalf("unexpected exchange response: %+v", exchanged)
	}

	refreshed, err := client.RefreshOAuthToken(context.Background(), OAuthTokenRequest{
		ClientID:     "patchxnote-local-dev",
		RefreshToken: oldRefresh,
	})
	if err != nil {
		t.Fatalf("refresh oauth token: %v", err)
	}
	if refreshed.RefreshToken == oldRefresh || refreshed.RefreshToken == "" {
		t.Fatalf("expected rotated refresh token, got %+v", refreshed)
	}

	if err := client.RevokeOAuthToken(context.Background(), oldRefresh); err != nil {
		t.Fatalf("revoke oauth token: %v", err)
	}
	if !metadataCalled || !exchangeCalled || !refreshCalled || !revokeCalled {
		t.Fatalf("expected all oauth calls metadata=%t exchange=%t refresh=%t revoke=%t", metadataCalled, exchangeCalled, refreshCalled, revokeCalled)
	}
}

func TestOAuthErrorDoesNotEchoTokenResponseBody(t *testing.T) {
	secretLike := strings.Repeat("s", 43)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"` + secretLike + `"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 0)
	_, err := client.ExchangeOAuthCode(context.Background(), OAuthTokenRequest{
		Code:         strings.Repeat("c", 43),
		RedirectURI:  "http://127.0.0.1:49152/callback",
		ClientID:     "patchxnote-local-dev",
		CodeVerifier: strings.Repeat("v", 43),
	})
	if err == nil {
		t.Fatal("expected oauth error")
	}
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *api.Error, got %T", err)
	}
	if apiErr.Code != "invalid_grant" {
		t.Fatalf("unexpected oauth error: %+v", apiErr)
	}
	if strings.Contains(err.Error(), secretLike) {
		t.Fatalf("oauth error string leaked response body: %v", err)
	}
}

func TestAgentSetupSessionMethods(t *testing.T) {
	var createCalled bool
	var getCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/setup-sessions":
			createCalled = true
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected create method: %s", r.Method)
			}
			if r.Header.Get("Idempotency-Key") != "idem_setup_fixture" {
				t.Fatalf("missing setup idempotency key: %s", r.Header.Get("Idempotency-Key"))
			}
			var request AgentSetupSessionCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatalf("decode setup request: %v", err)
			}
			if request.ClientID != "cursor" || request.Profile != "default" {
				t.Fatalf("unexpected setup request: %+v", request)
			}
			writeJSONResponse(t, w, http.StatusCreated, AgentSetupSessionCreated{
				SessionID:           "setup_fixture",
				Status:              "pending",
				UserCode:            "PXNOTE-1234",
				VerificationURI:     "https://patchxnote.example/setup",
				VerificationURIFull: "https://patchxnote.example/setup?code=PXNOTE-1234",
				ExpiresInSeconds:    600,
				PollIntervalSeconds: 2,
			})
		case "/v1/agent/setup-sessions/setup_fixture":
			getCalled = true
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected get method: %s", r.Method)
			}
			writeJSONResponse(t, w, http.StatusOK, AgentSetupSessionStatus{
				SessionID: "setup_fixture",
				Status:    "approved",
				Session: &AgentSessionResponse{
					AccessToken:             strings.Repeat("a", 32),
					AccessExpiresInSeconds:  3600,
					RefreshToken:            strings.Repeat("r", 43),
					RefreshExpiresInSeconds: 2592000,
					Account:                 CurrentAccount{ID: "acct_fixture", Status: "active"},
					Scopes:                  []string{"agent:account.read"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, 0)
	created, err := client.CreateAgentSetupSession(context.Background(), AgentSetupSessionCreateRequest{
		ClientID:   "cursor",
		ClientName: "Cursor",
		Profile:    "default",
	}, "idem_setup_fixture")
	if err != nil {
		t.Fatalf("create setup session: %v", err)
	}
	if created.SessionID != "setup_fixture" || created.UserCode == "" {
		t.Fatalf("unexpected created session: %+v", created)
	}
	status, err := client.GetAgentSetupSession(context.Background(), created.SessionID)
	if err != nil {
		t.Fatalf("get setup session: %v", err)
	}
	if status.Status != "approved" || status.Session == nil || status.Session.RefreshToken == "" {
		t.Fatalf("unexpected setup status: %+v", status)
	}
	if !createCalled || !getCalled {
		t.Fatalf("expected create/get calls create=%t get=%t", createCalled, getCalled)
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
		case "/v1/agent/memories/mem_fixture_delivery/delivery-document":
			if r.URL.Query().Get("platform") != "desktop" {
				t.Fatalf("unexpected delivery document query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_memory_delivery_document_success.json")
		case "/v1/agent/memories/mem_fixture_delivery/model-io":
			if r.URL.RawQuery != "" {
				t.Fatalf("expected optional platform to be omitted, got query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_memory_model_io_success.json")
		case "/v1/agent/model-runs/mrun_fixture_run/io-trace":
			if r.URL.Query().Get("platform") != "mobile" {
				t.Fatalf("unexpected io trace query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_model_run_io_trace_success.json")
		case "/v1/agent/model-runs/mrun_fixture_orphan/io-trace":
			writeFixture(t, w, http.StatusOK, "agent_model_run_io_trace_without_memory_success.json")
		case "/v1/agent/model-io-traces":
			query := r.URL.Query()
			if query.Get("platform") != "mobile" || query.Get("task_type") != "daily_digest" ||
				query.Get("state") != "completed" || query.Get("business_id") != "business-fixture-1" ||
				query.Get("date_from") != "2026-08-13T00:00:00Z" || query.Get("limit") != "10" ||
				query.Get("cursor") != "cursor_trace_page_1" {
				t.Fatalf("unexpected model io traces query: %s", r.URL.RawQuery)
			}
			writeFixture(t, w, http.StatusOK, "agent_model_io_traces_page_success.json")
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
	if page.NextCursor != "cursor_page_2" || len(page.Items) != 2 {
		t.Fatalf("unexpected memories page: %+v", page)
	}
	if page.Items[0].Source != "structured_result" ||
		page.Items[1].Source != "model_io_trace" ||
		page.Items[1].RequestID != "mrun_fixture_synthetic_1" ||
		page.Items[1].TaskType != "event_summary" ||
		page.Items[1].Title != "合成记录总结" ||
		page.Items[1].Summary == "" {
		t.Fatalf("unexpected memory optional fields: %+v", page.Items)
	}
	memory, err := client.GetMemory(context.Background(), credentialMaterial, "mobile", "mem_fixture_1")
	if err != nil {
		t.Fatalf("get memory: %v", err)
	}
	if memory.ID != "mem_fixture_1" || memory.SourceAvailability != "metadata_only" || memory.Source != "structured_result" {
		t.Fatalf("unexpected memory: %+v", memory)
	}
	delivery, err := client.GetMemoryDeliveryDocument(context.Background(), credentialMaterial, "desktop", "mem_fixture_delivery")
	if err != nil {
		t.Fatalf("delivery document: %v", err)
	}
	if delivery.Title != "合成会议纪要" || len(delivery.Sections) != 1 || delivery.Memory == nil ||
		delivery.Memory.Source != "structured_result" || delivery.Trace.RequestID != "mrun_fixture_delivery" {
		t.Fatalf("unexpected delivery document: %+v", delivery)
	}
	modelIO, err := client.GetMemoryModelIO(context.Background(), credentialMaterial, "", "mem_fixture_delivery")
	if err != nil {
		t.Fatalf("memory model io: %v", err)
	}
	if modelIO.SourceText == nil || modelIO.SourceText.Text == "" || modelIO.Memory == nil || len(modelIO.ProviderAttemptsJSON) == 0 {
		t.Fatalf("unexpected memory model io: %+v", modelIO)
	}
	runIO, err := client.GetModelRunIOTrace(context.Background(), credentialMaterial, "mobile", "mrun_fixture_run")
	if err != nil {
		t.Fatalf("run io trace: %v", err)
	}
	if runIO.Memory == nil || runIO.Trace.Platform != "mobile" {
		t.Fatalf("unexpected run io trace: %+v", runIO)
	}
	orphanIO, err := client.GetModelRunIOTrace(context.Background(), credentialMaterial, "", "mrun_fixture_orphan")
	if err != nil {
		t.Fatalf("orphan io trace: %v", err)
	}
	if orphanIO.Memory != nil || orphanIO.Trace.RequestID != "mrun_fixture_orphan" {
		t.Fatalf("unexpected orphan io trace: %+v", orphanIO)
	}
	traces, err := client.ListModelIOTraces(context.Background(), credentialMaterial, ListModelIOTracesParams{
		Platform:   "mobile",
		TaskType:   "daily_digest",
		State:      "completed",
		BusinessID: "business-fixture-1",
		DateFrom:   "2026-08-13T00:00:00Z",
		Limit:      10,
		Cursor:     "cursor_trace_page_1",
	})
	if err != nil {
		t.Fatalf("list model io traces: %v", err)
	}
	if traces.NextCursor != "cursor_trace_page_2" || len(traces.Items) != 2 ||
		traces.Items[0].RequestID != "mrun_fixture_trace_1" ||
		traces.Items[0].Memory == nil ||
		traces.Items[1].SourceTextAvailability != "not_applicable" ||
		traces.Items[0].FieldBytes.ProviderResponseJSON == 0 {
		t.Fatalf("unexpected model io traces page: %+v", traces)
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
	_, err = client.ListModelIOTraces(context.Background(), strings.Repeat("e", 32), ListModelIOTracesParams{})
	if err == nil {
		t.Fatal("expected model io trace platform validation error")
	}
	_, err = client.ListModelIOTraces(context.Background(), strings.Repeat("e", 32), ListModelIOTracesParams{Platform: "mobile", Limit: 51})
	if err == nil {
		t.Fatal("expected model io trace limit validation error")
	}
}

func TestDeliveryAndModelIOValidateInputs(t *testing.T) {
	client := newTestClient(t, "http://example.test", 0)
	if _, err := client.GetMemoryDeliveryDocument(context.Background(), "token", "web", "mem_fixture"); err == nil {
		t.Fatal("expected invalid platform error")
	}
	if _, err := client.GetMemoryModelIO(context.Background(), "token", "", ""); err == nil {
		t.Fatal("expected missing memory id error")
	}
	if _, err := client.GetModelRunIOTrace(context.Background(), "token", "", ""); err == nil {
		t.Fatal("expected missing request id error")
	}
}

func TestAgentDeliveryAndModelIOFixtureSensitiveValueScan(t *testing.T) {
	fixtures := []string{
		"agent_memory_delivery_document_success.json",
		"agent_memory_model_io_success.json",
		"agent_model_io_traces_page_success.json",
		"agent_model_run_io_trace_success.json",
		"agent_model_run_io_trace_without_memory_success.json",
	}
	disallowed := []string{
		"access_token",
		"refresh_token",
		"Bearer ",
		"otp",
		"tos_sk",
		"\"sk\"",
		"credential",
		"protocol_mac",
		"raw_audio",
		"Authorization",
		"13800138000",
	}
	for _, fixture := range fixtures {
		body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "api", fixture))
		if err != nil {
			t.Fatalf("read fixture %s: %v", fixture, err)
		}
		text := string(body)
		for _, needle := range disallowed {
			if strings.Contains(text, needle) {
				t.Fatalf("fixture %s contains disallowed marker %q", fixture, needle)
			}
		}
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

func serverURLForTest(r *http.Request) string {
	return "http://" + r.Host + "/patchnote-test-api"
}
