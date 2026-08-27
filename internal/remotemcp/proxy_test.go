package remotemcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyPreservesIDAndRefreshesOnceOnUnauthorized(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			if r.Header.Get("Authorization") != "Bearer old_token" {
				t.Fatalf("expected old token, got %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"req-1","error":{"code":-32001,"message":"auth required","data":{"code":"auth_required"}}}`))
			return
		}
		if r.Header.Get("Authorization") != "Bearer new_token" {
			t.Fatalf("expected new token, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"req-1","result":{"tools":[]}}`))
	}))
	defer server.Close()

	client, err := New(Options{ServerBaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	proxy, err := NewProxy(ProxyOptions{Client: client, TokenProvider: &fakeTokenProvider{access: "old_token", refresh: "new_token"}})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}
	var stdout bytes.Buffer
	if err := proxy.Serve(context.Background(), strings.NewReader(`{"jsonrpc":"2.0","id":"req-1","method":"tools/list"}`+"\n"), &stdout); err != nil {
		t.Fatalf("serve proxy: %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &response); err != nil {
		t.Fatalf("decode proxy response: %v\n%s", err, stdout.String())
	}
	if response["id"] != "req-1" || response["result"] == nil {
		t.Fatalf("unexpected proxy response: %+v", response)
	}
	if calls != 2 {
		t.Fatalf("expected retry after refresh, got %d calls", calls)
	}
}

func TestProxyMapsMalformedRemoteResponseToJSONRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	client, err := New(Options{ServerBaseURL: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	proxy, err := NewProxy(ProxyOptions{Client: client, TokenProvider: &fakeTokenProvider{access: "token"}})
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}
	var stdout bytes.Buffer
	if err := proxy.Serve(context.Background(), strings.NewReader(`{"jsonrpc":"2.0","id":7,"method":"initialize"}`+"\n"), &stdout); err != nil {
		t.Fatalf("serve proxy: %v", err)
	}
	var response struct {
		ID    int `json:"id"`
		Error struct {
			Data map[string]string `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &response); err != nil {
		t.Fatalf("decode error response: %v\n%s", err, stdout.String())
	}
	if response.ID != 7 || response.Error.Data["code"] != "remote_protocol_error" {
		t.Fatalf("unexpected error response: %+v", response)
	}
}

type fakeTokenProvider struct {
	access  string
	refresh string
}

func (f *fakeTokenProvider) AccessToken(ctx context.Context) (string, bool, error) {
	if f.access == "" {
		return "", false, nil
	}
	return f.access, true, nil
}

func (f *fakeTokenProvider) RefreshNow(ctx context.Context) (string, bool, error) {
	if f.refresh == "" {
		return "", false, nil
	}
	f.access = f.refresh
	return f.refresh, true, nil
}
