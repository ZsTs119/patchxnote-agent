package webhook

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const PatchXNoteWebhookPayloadVersion = "1"

type Message struct {
	Title    string            `json:"title"`
	Markdown string            `json:"markdown"`
	Memory   *MessageMemory    `json:"memory,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type MessageMemory struct {
	ID       string `json:"id,omitempty"`
	Platform string `json:"platform,omitempty"`
}

type RenderedRequest struct {
	URL         string
	ContentType string
	Body        []byte
}

func RenderRequest(target Target, webhookURL string, signingSecret string, message Message, now time.Time) (RenderedRequest, error) {
	if strings.TrimSpace(message.Markdown) == "" {
		return RenderedRequest{}, fmt.Errorf("webhook markdown is required")
	}
	if strings.TrimSpace(message.Title) == "" {
		message.Title = "PatchXNote 记录"
	}
	switch target.Type {
	case TargetTypeFeishu:
		return renderFeishuRequest(webhookURL, signingSecret, message, now)
	case TargetTypeDingTalk:
		return renderDingTalkRequest(webhookURL, signingSecret, message, now)
	case TargetTypeGeneric:
		return renderGenericRequest(webhookURL, message)
	default:
		return RenderedRequest{}, fmt.Errorf("unsupported webhook target type %q", target.Type)
	}
}

func renderFeishuRequest(webhookURL string, signingSecret string, message Message, now time.Time) (RenderedRequest, error) {
	cardMarkdown := feishuCardMarkdown(message.Markdown, message.Title)
	body := map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"config": map[string]any{
				"wide_screen_mode": true,
			},
			"header": map[string]any{
				"template": "blue",
				"title": map[string]any{
					"tag":     "plain_text",
					"content": message.Title,
				},
			},
			"elements": []any{
				map[string]any{
					"tag": "div",
					"text": map[string]any{
						"tag":     "lark_md",
						"content": cardMarkdown,
					},
				},
			},
		},
	}
	if strings.TrimSpace(signingSecret) != "" {
		timestamp := now.Unix()
		sign, err := FeishuSign(timestamp, signingSecret)
		if err != nil {
			return RenderedRequest{}, err
		}
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return RenderedRequest{}, fmt.Errorf("encode feishu payload: %w", err)
	}
	return RenderedRequest{URL: webhookURL, ContentType: "application/json", Body: encoded}, nil
}

func feishuCardMarkdown(markdown string, title string) string {
	lines := strings.Split(markdown, "\n")
	out := make([]string, 0, len(lines))
	inFence := false
	skippedTitle := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			out = append(out, line)
			continue
		}
		if !inFence {
			if heading, ok := markdownHeading(trimmed); ok {
				if !skippedTitle && strings.TrimSpace(heading) == strings.TrimSpace(title) {
					skippedTitle = true
					continue
				}
				out = append(out, "**"+heading+"**")
				continue
			}
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func markdownHeading(line string) (string, bool) {
	if line == "" || line[0] != '#' {
		return "", false
	}
	i := 0
	for i < len(line) && i < 6 && line[i] == '#' {
		i++
	}
	if i == 0 || i >= len(line) || line[i] != ' ' {
		return "", false
	}
	heading := strings.TrimSpace(line[i+1:])
	return heading, heading != ""
}

func renderDingTalkRequest(webhookURL string, signingSecret string, message Message, now time.Time) (RenderedRequest, error) {
	requestURL := webhookURL
	if strings.TrimSpace(signingSecret) != "" {
		parsed, err := url.Parse(webhookURL)
		if err != nil {
			return RenderedRequest{}, fmt.Errorf("parse dingtalk webhook URL: %w", err)
		}
		timestamp := now.UnixMilli()
		query := parsed.Query()
		query.Set("timestamp", fmt.Sprintf("%d", timestamp))
		query.Set("sign", DingTalkSign(timestamp, signingSecret))
		parsed.RawQuery = query.Encode()
		requestURL = parsed.String()
	}
	body := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"title": message.Title,
			"text":  message.Markdown,
		},
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return RenderedRequest{}, fmt.Errorf("encode dingtalk payload: %w", err)
	}
	return RenderedRequest{URL: requestURL, ContentType: "application/json", Body: encoded}, nil
}

func renderGenericRequest(webhookURL string, message Message) (RenderedRequest, error) {
	payload := map[string]any{
		"source":   "patchxnote",
		"version":  PatchXNoteWebhookPayloadVersion,
		"title":    message.Title,
		"markdown": message.Markdown,
	}
	if message.Memory != nil && (message.Memory.ID != "" || message.Memory.Platform != "") {
		payload["memory"] = message.Memory
	}
	metadata := map[string]string{"source": "file"}
	for key, value := range message.Metadata {
		metadata[key] = value
	}
	payload["metadata"] = metadata

	encoded, err := json.Marshal(payload)
	if err != nil {
		return RenderedRequest{}, fmt.Errorf("encode generic payload: %w", err)
	}
	return RenderedRequest{URL: webhookURL, ContentType: "application/json", Body: encoded}, nil
}
