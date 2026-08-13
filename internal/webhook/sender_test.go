package webhook

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSenderSuccessAndProviderFailures(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		switch r.URL.Path {
		case "/feishu-ok":
			fmt.Fprint(w, `{"code":0,"msg":"success"}`)
		case "/feishu-fail":
			fmt.Fprint(w, `{"code":19001,"msg":"keyword missing"}`)
		case "/generic-ok":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	targets := []ResolvedTarget{
		resolvedTarget("飞书成功", TargetTypeFeishu, server.URL+"/feishu-ok"),
		resolvedTarget("飞书失败", TargetTypeFeishu, server.URL+"/feishu-fail"),
		resolvedTarget("通用成功", TargetTypeGeneric, server.URL+"/generic-ok"),
	}
	results, err := NewSender(server.Client()).Send(context.Background(), targets, fixtureMessage(), SendOptions{})
	if err == nil {
		t.Fatal("expected mixed failure")
	}
	if len(results) != 3 || !results[0].Success || results[1].Success || !results[2].Success {
		t.Fatalf("unexpected mixed results: %+v", results)
	}
	if results[1].ProviderCode != "19001" || !strings.Contains(results[1].ProviderMessage, "keyword") {
		t.Fatalf("unexpected provider failure: %+v", results[1])
	}
	if strings.Join(calls, ",") != "/feishu-ok,/feishu-fail,/generic-ok" {
		t.Fatalf("expected sequential calls, got %v", calls)
	}
}

func TestSenderDingTalkProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"errcode":310000,"errmsg":"keywords not in content"}`)
	}))
	defer server.Close()

	results, err := NewSender(server.Client()).Send(context.Background(), []ResolvedTarget{
		resolvedTarget("钉钉", TargetTypeDingTalk, server.URL),
	}, fixtureMessage(), SendOptions{})
	if err == nil {
		t.Fatal("expected dingtalk provider error")
	}
	if len(results) != 1 || results[0].ProviderCode != "310000" || results[0].Error == "" {
		t.Fatalf("unexpected result: %+v", results)
	}
}

func TestSenderRedirectHTTP429InvalidJSONAndSecretFreeErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/rate":
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, strings.Repeat("x", maxProviderBody+20))
		case "/invalid":
			fmt.Fprint(w, `not-json`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	targets := []ResolvedTarget{
		resolvedTarget("跳转", TargetTypeGeneric, server.URL+"/redirect"),
		resolvedTarget("限流", TargetTypeGeneric, server.URL+"/rate"),
		resolvedTarget("非法", TargetTypeFeishu, server.URL+"/invalid"),
	}
	results, err := NewSender(server.Client()).Send(context.Background(), targets, fixtureMessage(), SendOptions{})
	if err == nil {
		t.Fatal("expected failures")
	}
	if results[0].StatusCode != http.StatusFound || results[0].Success {
		t.Fatalf("redirect should be surfaced, got %+v", results[0])
	}
	if results[1].StatusCode != http.StatusTooManyRequests || !strings.Contains(results[1].ProviderMessage, "truncated") {
		t.Fatalf("rate limit should be surfaced with bounded excerpt, got %+v", results[1])
	}
	if !strings.Contains(results[2].Error, "decode feishu") {
		t.Fatalf("invalid json should fail decode, got %+v", results[2])
	}
	for _, result := range results {
		if strings.Contains(result.Error+result.ProviderMessage, server.URL) {
			t.Fatalf("result leaked raw URL: %+v", result)
		}
	}
}

func TestSenderDuplicateTargetFailsBeforeRequest(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()

	_, err := NewSender(server.Client()).Send(context.Background(), []ResolvedTarget{
		resolvedTarget("同名", TargetTypeGeneric, server.URL),
		resolvedTarget("同名", TargetTypeGeneric, server.URL),
	}, fixtureMessage(), SendOptions{})
	if err == nil {
		t.Fatal("expected duplicate target error")
	}
	if calls != 0 {
		t.Fatalf("expected no request before duplicate validation, got %d", calls)
	}
}

func TestSenderEmptyTargetFails(t *testing.T) {
	_, err := NewSender(nil).Send(context.Background(), nil, fixtureMessage(), SendOptions{})
	if err == nil {
		t.Fatal("expected empty target error")
	}
}

func TestSenderTimeoutIsClear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	results, err := NewSender(server.Client()).Send(context.Background(), []ResolvedTarget{
		resolvedTarget("慢请求", TargetTypeGeneric, server.URL),
	}, fixtureMessage(), SendOptions{Timeout: time.Millisecond})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if len(results) != 1 || !strings.Contains(results[0].Error, "timed out") {
		t.Fatalf("unexpected timeout result: %+v", results)
	}
}

func resolvedTarget(alias string, targetType TargetType, rawURL string) ResolvedTarget {
	return ResolvedTarget{
		Target: Target{
			Alias:     alias,
			Type:      targetType,
			Enabled:   true,
			MaskedURL: MaskWebhookURL(rawURL),
		},
		URL: rawURL,
	}
}
