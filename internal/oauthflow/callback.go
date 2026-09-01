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
	title := "登录未完成"
	message := "请回到应用重新打开登录。"
	note := "没有保存新的登录信息。"
	badge := "未完成"
	badgeClass := "state-badge error"
	if success {
		title = "登录已完成"
		message = "可以回到编辑器继续使用。"
		note = "此页面可以关闭。"
		badge = "已完成"
		badgeClass = "state-badge success"
	}
	_, _ = w.Write([]byte(fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <style>
    :root {
      color-scheme: light;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", Arial, sans-serif;
      background: #f6f7f9;
      color: #111827;
    }
    * {
      box-sizing: border-box;
    }
    body {
      margin: 0;
      min-height: 100dvh;
      display: grid;
      place-items: center;
      padding: 24px;
      background: #f6f7f9;
    }
    main {
      width: min(420px, 100%%);
      border: 1px solid #dde3ee;
      border-radius: 8px;
      background: #ffffff;
      box-shadow: 0 24px 70px rgba(15, 23, 42, 0.12);
      padding: 32px;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 26px;
      color: #111827;
      font-size: 15px;
      font-weight: 700;
    }
    .brand-mark {
      display: inline-grid;
      width: 32px;
      height: 32px;
      place-items: center;
      border-radius: 8px;
      background: #111827;
      color: #ffffff;
      font-size: 13px;
      font-weight: 800;
    }
    .state-badge {
      display: inline-flex;
      align-items: center;
      height: 28px;
      border-radius: 8px;
      padding: 0 10px;
      font-size: 13px;
      font-weight: 700;
    }
    .state-badge.success {
      background: #ecfdf5;
      color: #067647;
    }
    .state-badge.error {
      background: #fef3f2;
      color: #b42318;
    }
    h1 {
      margin: 16px 0 10px;
      color: #111827;
      font-size: 28px;
      line-height: 1.2;
      letter-spacing: 0;
    }
    p {
      margin: 0;
      color: #5b6472;
      line-height: 1.6;
    }
    .note {
      margin-top: 18px;
      font-size: 14px;
    }
    @media (max-width: 520px) {
      body {
        padding: 14px;
      }
      main {
        padding: 24px;
      }
    }
  </style>
</head>
<body>
  <main>
    <div class="brand"><span class="brand-mark">PX</span><span>PatchXNote</span></div>
    <div class="%s">%s</div>
    <h1>%s</h1>
    <p>%s</p>
    <p class="note">%s</p>
  </main>
</body>
</html>`, title, badgeClass, badge, title, message, note)))
}
