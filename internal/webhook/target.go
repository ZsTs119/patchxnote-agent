package webhook

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type TargetType string

const (
	TargetTypeFeishu   TargetType = "feishu"
	TargetTypeDingTalk TargetType = "dingtalk"
	TargetTypeGeneric  TargetType = "generic"
)

type Target struct {
	Alias     string     `json:"alias"`
	Type      TargetType `json:"type"`
	Enabled   bool       `json:"enabled"`
	MaskedURL string     `json:"masked_url"`
	Template  string     `json:"template,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func ValidateAlias(alias string) (string, error) {
	normalized := strings.TrimSpace(alias)
	if normalized == "" {
		return "", fmt.Errorf("webhook alias is required")
	}
	if utf8.RuneCountInString(normalized) > 64 {
		return "", fmt.Errorf("webhook alias must be at most 64 characters")
	}
	for _, r := range normalized {
		if unicode.IsControl(r) || r == '\t' || r == '\n' || r == '\r' {
			return "", fmt.Errorf("webhook alias must not contain control characters")
		}
		switch r {
		case '/', '\\':
			return "", fmt.Errorf("webhook alias must not contain path separators")
		}
	}
	return normalized, nil
}

func ValidateType(value string) (TargetType, error) {
	switch TargetType(strings.ToLower(strings.TrimSpace(value))) {
	case TargetTypeFeishu:
		return TargetTypeFeishu, nil
	case TargetTypeDingTalk:
		return TargetTypeDingTalk, nil
	case TargetTypeGeneric:
		return TargetTypeGeneric, nil
	default:
		return "", fmt.Errorf("webhook type must be feishu, dingtalk, or generic")
	}
}

func ValidateWebhookURL(value string) (string, error) {
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("webhook URL must not contain control characters")
		}
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("webhook URL is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse webhook URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("webhook URL must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("webhook URL host is required")
	}
	if parsed.Fragment != "" {
		return "", fmt.Errorf("webhook URL fragments are not supported")
	}
	return parsed.String(), nil
}

func MaskWebhookURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "<invalid>"
	}
	masked := *parsed
	if masked.User != nil {
		masked.User = url.User("***")
	}
	if masked.RawQuery != "" {
		query := masked.Query()
		for key := range query {
			query.Set(key, "***")
		}
		masked.RawQuery = query.Encode()
	}
	parts := strings.Split(masked.Path, "/")
	lastPathSegment := -1
	for i, part := range parts {
		if part != "" {
			lastPathSegment = i
		}
	}
	for i, part := range parts {
		if looksSecretPathSegment(part) || i == lastPathSegment {
			parts[i] = maskToken(part)
		}
	}
	masked.Path = strings.Join(parts, "/")
	return masked.String()
}

func looksSecretPathSegment(segment string) bool {
	if len(segment) < 16 {
		return false
	}
	var alphaNum int
	for _, r := range segment {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alphaNum++
		}
	}
	return alphaNum >= 12
}

func maskToken(value string) string {
	runes := []rune(value)
	if len(runes) <= 8 {
		return "***"
	}
	return string(runes[:4]) + "..." + string(runes[len(runes)-4:])
}
