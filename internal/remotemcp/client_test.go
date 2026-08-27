package remotemcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientPostsJSONRPCWithBearerToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/patchnote-test-api/mcp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer "+strings.Repeat("a", 43) {
			t.Fatal("missing bearer token")
		}
		if !strings.Contains(r.Header.Get("Accept"), "application/json") || !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			t.Fatalf("unexpected accept header: %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`))
	}))
	defer server.Close()

	client, err := New(Options{ServerBaseURL: server.URL + "/patchnote-test-api"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	response, err := client.Do(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`), strings.Repeat("a", 43))
	if err != nil {
		t.Fatalf("do remote mcp: %v", err)
	}
	if response.StatusCode != http.StatusOK || response.NoResponse {
		t.Fatalf("unexpected response: %+v", response)
	}
	if err := validJSONRPCResponse(response.Body, json.RawMessage("1")); err != nil {
		t.Fatalf("validate response: %v", err)
	}
}

func TestClientRejectsEventStreamUntilSupported(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {}\n\n"))
	}))
	defer server.Close()

	client, err := New(Options{ServerBaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Do(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`), "")
	if err == nil || !strings.Contains(err.Error(), "streamable HTTP") {
		t.Fatalf("expected streamable HTTP compatibility error, got %v", err)
	}
}
