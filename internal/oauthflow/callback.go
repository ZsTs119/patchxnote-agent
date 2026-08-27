package oauthflow

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrCallbackStateMismatch = errors.New("oauth callback state mismatch")
	ErrCallbackMissingCode   = errors.New("oauth callback missing code")
	ErrCallbackDenied        = errors.New("oauth authorization was denied")
)

type CallbackResult struct {
	Code             string
	State            string
	Error            string
	ErrorDescription string
}

type CallbackServer struct {
	listener net.Listener
	server   *http.Server
	done     chan CallbackResult
	once     sync.Once
}

func StartCallbackServer(ctx context.Context, expectedState string) (*CallbackServer, error) {
	expectedState = strings.TrimSpace(expectedState)
	if expectedState == "" {
		return nil, fmt.Errorf("oauth callback state is required")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start oauth callback listener: %w", err)
	}
	callback := &CallbackServer{
		listener: listener,
		done:     make(chan CallbackResult, 1),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", callback.handle(expectedState))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeCallbackHTML(w, http.StatusNotFound, false)
	})
	callback.server = &http.Server{Handler: mux}
	go func() {
		_ = callback.server.Serve(listener)
	}()
	go func() {
		<-ctx.Done()
		_ = callback.Close(context.Background())
	}()
	return callback, nil
}

func (s *CallbackServer) RedirectURI() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return "http://" + s.listener.Addr().String() + "/callback"
}

func (s *CallbackServer) Wait(ctx context.Context) (CallbackResult, error) {
	if s == nil {
		return CallbackResult{}, fmt.Errorf("oauth callback server is not running")
	}
	select {
	case result := <-s.done:
		if result.Error != "" {
			return result, ErrCallbackDenied
		}
		if strings.TrimSpace(result.Code) == "" {
			return result, ErrCallbackMissingCode
		}
		return result, nil
	case <-ctx.Done():
		_ = s.Close(context.Background())
		return CallbackResult{}, fmt.Errorf("oauth callback timed out or was cancelled")
	}
}

func (s *CallbackServer) Close(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *CallbackServer) handle(expectedState string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeCallbackHTML(w, http.StatusMethodNotAllowed, false)
			return
		}
		query := r.URL.Query()
		result := CallbackResult{
			Code:             strings.TrimSpace(query.Get("code")),
			State:            strings.TrimSpace(query.Get("state")),
			Error:            strings.TrimSpace(query.Get("error")),
			ErrorDescription: strings.TrimSpace(query.Get("error_description")),
		}
		success := false
		status := http.StatusBadRequest
		switch {
		case result.Error != "":
		case result.State != expectedState:
			result.Error = "state_mismatch"
		case result.Code == "":
			result.Error = "missing_code"
		default:
			success = true
			status = http.StatusOK
		}
		s.complete(result)
		writeCallbackHTML(w, status, success)
	}
}

func (s *CallbackServer) complete(result CallbackResult) {
	s.once.Do(func() {
		s.done <- result
		close(s.done)
		go func() {
			_ = s.Close(context.Background())
		}()
	})
}

func writeCallbackHTML(w http.ResponseWriter, status int, success bool) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if success {
		_, _ = w.Write([]byte(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>PatchXNote MCP 授权成功</title></head><body><h1>PatchXNote MCP 授权成功</h1><p>可以回到编辑器继续使用。</p></body></html>`))
		return
	}
	_, _ = w.Write([]byte(`<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>PatchXNote MCP 授权未完成</title></head><body><h1>PatchXNote MCP 授权未完成</h1><p>请回到终端重新发起登录。</p></body></html>`))
}
