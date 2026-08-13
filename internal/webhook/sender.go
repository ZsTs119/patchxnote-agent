package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultSendTimeout = 15 * time.Second
	maxProviderBody    = 512
)

type ResolvedTarget struct {
	Target        Target
	URL           string
	SigningSecret string
}

type Sender struct {
	httpClient *http.Client
	now        func() time.Time
}

type SendOptions struct {
	Timeout time.Duration
}

type SendResult struct {
	Alias           string `json:"alias"`
	Type            string `json:"type"`
	Success         bool   `json:"success"`
	StatusCode      int    `json:"status_code,omitempty"`
	ProviderCode    string `json:"provider_code,omitempty"`
	ProviderMessage string `json:"provider_message,omitempty"`
	MaskedURL       string `json:"masked_url,omitempty"`
	Error           string `json:"error,omitempty"`
}

func NewSender(httpClient *http.Client) *Sender {
	return &Sender{
		httpClient: httpClient,
		now:        time.Now,
	}
}

func (s *Sender) Send(ctx context.Context, targets []ResolvedTarget, message Message, options SendOptions) ([]SendResult, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one webhook target is required")
	}
	if err := validateSendTargets(targets); err != nil {
		return nil, err
	}
	client := s.client(options.Timeout)
	results := make([]SendResult, 0, len(targets))
	for _, target := range targets {
		result := s.sendOne(ctx, client, target, message)
		results = append(results, result)
	}
	var failed int
	for _, result := range results {
		if !result.Success {
			failed++
		}
	}
	if failed > 0 {
		return results, fmt.Errorf("%d webhook target(s) failed", failed)
	}
	return results, nil
}

func (s *Sender) client(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = DefaultSendTimeout
	}
	if s.httpClient != nil {
		client := *s.httpClient
		if client.Timeout == 0 {
			client.Timeout = timeout
		}
		client.CheckRedirect = noRedirect
		return &client
	}
	return &http.Client{Timeout: timeout, CheckRedirect: noRedirect}
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func (s *Sender) sendOne(ctx context.Context, client *http.Client, target ResolvedTarget, message Message) SendResult {
	result := SendResult{
		Alias:     target.Target.Alias,
		Type:      string(target.Target.Type),
		MaskedURL: target.Target.MaskedURL,
	}
	rendered, err := RenderRequest(target.Target, target.URL, target.SigningSecret, message, s.now().UTC())
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rendered.URL, bytes.NewReader(rendered.Body))
	if err != nil {
		result.Error = "build webhook request failed"
		return result
	}
	req.Header.Set("Content-Type", rendered.ContentType)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = classifySendError(err)
		return result
	}
	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxProviderBody+1))
	if readErr != nil {
		result.Error = "read provider response failed"
		return result
	}
	excerpt := string(body)
	if len(body) > maxProviderBody {
		excerpt = string(body[:maxProviderBody]) + "...(truncated)"
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Error = "provider HTTP error"
		result.ProviderMessage = strings.TrimSpace(excerpt)
		return result
	}

	switch target.Target.Type {
	case TargetTypeGeneric:
		result.Success = true
		return result
	case TargetTypeFeishu:
		return parseFeishuResult(result, body)
	case TargetTypeDingTalk:
		return parseDingTalkResult(result, body)
	default:
		result.Error = "unsupported webhook target type"
		return result
	}
}

func validateSendTargets(targets []ResolvedTarget) error {
	seen := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		alias, err := ValidateAlias(target.Target.Alias)
		if err != nil {
			return err
		}
		if _, ok := seen[alias]; ok {
			return fmt.Errorf("duplicate webhook target alias %q", alias)
		}
		seen[alias] = struct{}{}
	}
	return nil
}

func classifySendError(err error) string {
	if errors.Is(err, context.Canceled) {
		return "webhook request canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "webhook request timed out"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "webhook request timed out"
	}
	return "send webhook request failed"
}

func parseFeishuResult(result SendResult, body []byte) SendResult {
	if len(strings.TrimSpace(string(body))) == 0 {
		result.Success = true
		return result
	}
	var response struct {
		Code          *int   `json:"code,omitempty"`
		Msg           string `json:"msg,omitempty"`
		StatusCode    *int   `json:"StatusCode,omitempty"`
		StatusMessage string `json:"StatusMessage,omitempty"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		result.Error = "decode feishu response failed"
		result.ProviderMessage = safeBodyExcerpt(body)
		return result
	}
	code := 0
	if response.Code != nil {
		code = *response.Code
	} else if response.StatusCode != nil {
		code = *response.StatusCode
	}
	message := response.Msg
	if message == "" {
		message = response.StatusMessage
	}
	if code != 0 {
		result.ProviderCode = fmt.Sprintf("%d", code)
		result.ProviderMessage = message
		result.Error = "feishu provider error"
		return result
	}
	result.Success = true
	result.ProviderMessage = message
	return result
}

func parseDingTalkResult(result SendResult, body []byte) SendResult {
	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		result.Error = "decode dingtalk response failed"
		result.ProviderMessage = safeBodyExcerpt(body)
		return result
	}
	if response.ErrCode != 0 {
		result.ProviderCode = fmt.Sprintf("%d", response.ErrCode)
		result.ProviderMessage = response.ErrMsg
		result.Error = "dingtalk provider error"
		return result
	}
	result.Success = true
	result.ProviderMessage = response.ErrMsg
	return result
}

func safeBodyExcerpt(body []byte) string {
	if len(body) > maxProviderBody {
		return string(body[:maxProviderBody]) + "...(truncated)"
	}
	return strings.TrimSpace(string(body))
}
