package modelio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/localfile"
)

const DefaultInlineLimitBytes = 16 * 1024

type Field string

const (
	FieldSourceText       Field = "source-text"
	FieldProviderResponse Field = "provider-response"
	FieldParsedResult     Field = "parsed-result"
	FieldPackagedResult   Field = "packaged-result"
)

type Lookup struct {
	MemoryID  string
	RequestID string
	Platform  string
}

func (l Lookup) Kind() string {
	if strings.TrimSpace(l.MemoryID) != "" {
		return "memory"
	}
	if strings.TrimSpace(l.RequestID) != "" {
		return "request"
	}
	return ""
}

type FieldResult struct {
	Field      Field
	Available  bool
	Status     string
	MediaType  string
	Text       string
	RawJSON    json.RawMessage
	OutputPath string
	Written    bool
	LookupKind string
	MemoryID   string
	RequestID  string
	Platform   string
	Source     string
	Memory     *api.AgentDeliveryMemory
	Trace      api.AgentDeliveryTrace
}

func ValidateField(value string) (Field, error) {
	field := Field(strings.TrimSpace(value))
	switch field {
	case FieldSourceText, FieldProviderResponse, FieldParsedResult, FieldPackagedResult:
		return field, nil
	default:
		return "", fmt.Errorf("field must be source-text, provider-response, parsed-result, or packaged-result")
	}
}

func ValidateLookup(lookup Lookup) error {
	memoryID := strings.TrimSpace(lookup.MemoryID)
	requestID := strings.TrimSpace(lookup.RequestID)
	if (memoryID == "") == (requestID == "") {
		return fmt.Errorf("exactly one of memory_id or request_id is required")
	}
	platform := strings.TrimSpace(lookup.Platform)
	if platform != "" && platform != "mobile" && platform != "desktop" {
		return fmt.Errorf("platform must be mobile or desktop")
	}
	return nil
}

func SelectField(export api.AgentModelIOExport, lookup Lookup, field Field) (FieldResult, error) {
	if _, err := ValidateField(string(field)); err != nil {
		return FieldResult{}, err
	}
	if err := ValidateLookup(lookup); err != nil {
		return FieldResult{}, err
	}
	result := FieldResult{
		Field:      field,
		LookupKind: lookup.Kind(),
		MemoryID:   strings.TrimSpace(lookup.MemoryID),
		RequestID:  strings.TrimSpace(lookup.RequestID),
		Platform:   resolvedPlatform(lookup, export),
		Source:     export.Source,
		Memory:     export.Memory,
		Trace:      export.Trace,
	}
	switch field {
	case FieldSourceText:
		return selectSourceText(export, result), nil
	case FieldProviderResponse:
		return selectJSONField(export.ProviderResponseJSON, export.FieldStatus.ProviderResponseJSON, result), nil
	case FieldParsedResult:
		return selectJSONField(export.ParsedResultJSON, export.FieldStatus.ParsedResultJSON, result), nil
	case FieldPackagedResult:
		return selectJSONField(export.PackagedResultJSON, export.FieldStatus.PackagedResultJSON, result), nil
	default:
		return FieldResult{}, fmt.Errorf("unsupported model IO field %q", field)
	}
}

func selectSourceText(export api.AgentModelIOExport, result FieldResult) FieldResult {
	result.MediaType = "text/plain"
	if export.SourceText == nil {
		result.Status = "missing"
		return result
	}
	status := strings.TrimSpace(export.SourceText.Availability)
	text := export.SourceText.Text
	if status == "" {
		if strings.TrimSpace(text) == "" {
			status = "missing"
		} else {
			status = "available"
		}
	}
	result.Status = status
	if status != "available" {
		return result
	}
	if text == "" {
		result.Status = "empty"
		return result
	}
	result.Available = true
	result.Text = text
	return result
}

func selectJSONField(raw json.RawMessage, status string, result FieldResult) FieldResult {
	result.MediaType = "application/json"
	trimmed := bytes.TrimSpace(raw)
	status = strings.TrimSpace(status)
	if status == "" {
		if len(trimmed) == 0 {
			status = "missing"
		} else {
			status = "available"
		}
	}
	result.Status = status
	if len(trimmed) == 0 {
		if status == "available" {
			result.Status = "empty"
		}
		return result
	}
	if bytes.Equal(trimmed, []byte("null")) {
		if status == "available" {
			result.Status = "null"
		}
		return result
	}
	if !json.Valid(trimmed) {
		result.Status = "invalid_json"
		return result
	}
	if status != "available" {
		return result
	}
	result.Available = true
	result.RawJSON = append(json.RawMessage(nil), trimmed...)
	return result
}

func (r FieldResult) ContentBytes() ([]byte, error) {
	if !r.Available {
		return nil, nil
	}
	switch r.MediaType {
	case "text/plain":
		body := []byte(r.Text)
		if len(body) == 0 || body[len(body)-1] != '\n' {
			body = append(body, '\n')
		}
		return body, nil
	case "application/json":
		return PrettyJSON(r.RawJSON)
	default:
		return nil, fmt.Errorf("unsupported field media type %q", r.MediaType)
	}
}

func PrettyJSON(raw json.RawMessage) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("JSON content is empty")
	}
	if !json.Valid(trimmed) {
		return nil, fmt.Errorf("JSON content is invalid")
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, trimmed, "", "  "); err != nil {
		return nil, fmt.Errorf("format JSON content: %w", err)
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func (r FieldResult) InlineValue() (any, error) {
	if !r.Available {
		return nil, nil
	}
	if r.MediaType == "text/plain" {
		return r.Text, nil
	}
	var value any
	if err := json.Unmarshal(r.RawJSON, &value); err != nil {
		return nil, fmt.Errorf("decode JSON content: %w", err)
	}
	return value, nil
}

func (r FieldResult) Summary(includeContent bool) (map[string]any, error) {
	result := map[string]any{
		"field":       string(r.Field),
		"available":   r.Available,
		"status":      r.Status,
		"media_type":  r.MediaType,
		"lookup_kind": r.LookupKind,
		"written":     r.Written,
	}
	if r.MemoryID != "" {
		result["memory_id"] = r.MemoryID
	}
	if r.RequestID != "" {
		result["request_id"] = r.RequestID
	}
	if r.Platform != "" {
		result["platform"] = r.Platform
	}
	if r.Source != "" {
		result["source"] = r.Source
	}
	if r.Memory != nil {
		result["memory"] = r.Memory
	}
	if r.Trace.RequestID != "" || r.Trace.TraceID != "" {
		result["trace"] = r.Trace
	}
	if r.OutputPath != "" {
		result["out"] = r.OutputPath
	}
	if includeContent && r.Available {
		value, err := r.InlineValue()
		if err != nil {
			return nil, err
		}
		if r.MediaType == "text/plain" {
			result["text"] = value
		} else {
			result["json"] = value
		}
	}
	return result, nil
}

func WriteFieldFile(path string, result FieldResult, force bool) (FieldResult, error) {
	if !result.Available {
		return result, nil
	}
	abs, err := localfile.AbsPath(path)
	if err != nil {
		return result, err
	}
	body, err := result.ContentBytes()
	if err != nil {
		return result, err
	}
	if err := localfile.AtomicWriteFile(abs, body, 0o600, force); err != nil {
		return result, err
	}
	result.OutputPath = abs
	result.Written = true
	return result, nil
}

func WriteExportFile(path string, export api.AgentModelIOExport, force bool) (string, error) {
	abs, err := localfile.AbsPath(path)
	if err != nil {
		return "", err
	}
	body, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode model IO export: %w", err)
	}
	body = append(body, '\n')
	if err := localfile.AtomicWriteFile(abs, body, 0o600, force); err != nil {
		return "", err
	}
	return abs, nil
}

func resolvedPlatform(lookup Lookup, export api.AgentModelIOExport) string {
	if strings.TrimSpace(lookup.Platform) != "" {
		return strings.TrimSpace(lookup.Platform)
	}
	if export.Memory != nil && export.Memory.Platform != "" {
		return export.Memory.Platform
	}
	return export.Trace.Platform
}
