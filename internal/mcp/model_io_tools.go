package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/modelio"
)

func defaultModelIOTools(server *Server) []Tool {
	return []Tool{
		modelIOTraceListTool(server),
		modelIOFieldTool(server, "patchxnote_get_model_io_source_text", "Get the explicit PatchXNote Agent source text/safe transcript projection.", modelio.FieldSourceText),
		modelIOFieldTool(server, "patchxnote_get_model_io_provider_response", "Get the explicit model provider response JSON for one PatchXNote Agent model IO trace.", modelio.FieldProviderResponse),
		modelIOFieldTool(server, "patchxnote_get_model_io_parsed_result", "Get the parsed model result JSON for one PatchXNote Agent model IO trace.", modelio.FieldParsedResult),
		modelIOFieldTool(server, "patchxnote_get_model_io_packaged_result", "Get the packaged structured result JSON for one PatchXNote Agent model IO trace.", modelio.FieldPackagedResult),
	}
}

func modelIOTraceListTool(server *Server) Tool {
	return Tool{
		Name:        "patchxnote_list_model_io_traces",
		Description: "List lightweight PatchXNote Agent model IO trace metadata and request IDs for one platform.",
		InputSchema: objectSchema(map[string]any{
			"platform":     platformProperty(),
			"request_id":   stringProperty(1, 160),
			"task_type":    stringProperty(1, 64),
			"state":        stringProperty(1, 64),
			"recording_id": stringProperty(1, 128),
			"event_id":     stringProperty(1, 128),
			"business_id":  stringProperty(1, 128),
			"date_from":    stringProperty(1, 64),
			"date_to":      stringProperty(1, 64),
			"limit":        integerProperty(1, 50),
			"cursor":       stringProperty(1, 512),
		}, []string{"platform"}),
		Annotations: readOnlyAnnotations(),
		handler:     server.handleListModelIOTraces,
		validator:   validateListModelIOTracesArgs,
	}
}

func modelIOFieldTool(server *Server, name string, description string, field modelio.Field) Tool {
	return Tool{
		Name:        name,
		Description: description,
		InputSchema: objectSchema(map[string]any{
			"memory_id":  stringProperty(1, 160),
			"request_id": stringProperty(1, 160),
			"platform":   platformProperty(),
			"out":        stringProperty(1, 4096),
			"force":      booleanProperty(),
		}, nil),
		Annotations: localWriteAnnotations(),
		handler: func(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
			return server.handleModelIOField(ctx, arguments, field)
		},
		validator: validateModelIOFieldArgs,
	}
}

func (s *Server) handleListModelIOTraces(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	params, err := decodeListModelIOTracesArgs(arguments)
	if err != nil {
		return CallToolResult{}, err
	}
	var page api.AgentModelIOTracePage
	err = s.withAgentAccessToken(ctx, params.Platform, func(accessToken string) error {
		var err error
		page, err = s.api.ListModelIOTraces(ctx, accessToken, params)
		return err
	})
	if err != nil {
		return CallToolResult{}, friendlyAgentMCPError(err)
	}
	return jsonResult(page)
}

func (s *Server) handleModelIOField(ctx context.Context, arguments json.RawMessage, field modelio.Field) (CallToolResult, error) {
	args, err := decodeModelIOFieldArgs(arguments)
	if err != nil {
		return CallToolResult{}, err
	}
	lookup := modelio.Lookup{MemoryID: args.MemoryID, RequestID: args.RequestID, Platform: args.Platform}
	export, err := s.fetchModelIOExport(ctx, lookup)
	if err != nil {
		return CallToolResult{}, err
	}
	result, err := modelio.SelectField(export, lookup, field)
	if err != nil {
		return CallToolResult{}, err
	}
	includeContent := false
	if args.Out != "" {
		result, err = modelio.WriteFieldFile(args.Out, result, args.Force)
		if err != nil {
			return CallToolResult{}, err
		}
	} else if result.Available {
		body, err := result.ContentBytes()
		if err != nil {
			return CallToolResult{}, err
		}
		if len(body) > maxToolOutputBytes {
			return CallToolResult{}, rpcErr(codeToolError, "output_too_large", "model IO field exceeds MCP output limit; pass out to write it to a file")
		}
		includeContent = true
	}
	summary, err := result.Summary(includeContent)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(summary)
}

func (s *Server) fetchModelIOExport(ctx context.Context, lookup modelio.Lookup) (api.AgentModelIOExport, error) {
	if lookup.MemoryID != "" {
		return s.fetchMemoryModelIO(ctx, lookup.Platform, lookup.MemoryID)
	}
	return s.fetchModelRunIOTrace(ctx, lookup.Platform, lookup.RequestID)
}

type modelIOFieldArgs struct {
	MemoryID  string `json:"memory_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Out       string `json:"out,omitempty"`
	Force     bool   `json:"force,omitempty"`
}

type modelIOTraceListArgs struct {
	Platform    string `json:"platform"`
	RequestID   string `json:"request_id,omitempty"`
	TaskType    string `json:"task_type,omitempty"`
	State       string `json:"state,omitempty"`
	RecordingID string `json:"recording_id,omitempty"`
	EventID     string `json:"event_id,omitempty"`
	BusinessID  string `json:"business_id,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
}

func decodeListModelIOTracesArgs(raw json.RawMessage) (api.ListModelIOTracesParams, error) {
	var args modelIOTraceListArgs
	if err := json.Unmarshal(argumentsOrEmpty(raw), &args); err != nil {
		return api.ListModelIOTracesParams{}, err
	}
	if args.Limit == 0 {
		args.Limit = 20
	}
	return api.ListModelIOTracesParams{
		Platform: args.Platform, RequestID: args.RequestID, TaskType: args.TaskType, State: args.State,
		RecordingID: args.RecordingID, EventID: args.EventID, BusinessID: args.BusinessID,
		DateFrom: args.DateFrom, DateTo: args.DateTo, Limit: args.Limit, Cursor: args.Cursor,
	}, nil
}

func decodeModelIOFieldArgs(raw json.RawMessage) (modelIOFieldArgs, error) {
	var args modelIOFieldArgs
	if err := json.Unmarshal(argumentsOrEmpty(raw), &args); err != nil {
		return modelIOFieldArgs{}, err
	}
	return args, nil
}

func validateListModelIOTracesArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := requirePlatform(args); err != nil {
		return err
	}
	if err := optionalString(args, "request_id", 1, 160); err != nil {
		return err
	}
	if err := optionalModelTaskType(args, "task_type"); err != nil {
		return err
	}
	if err := optionalModelTraceState(args, "state"); err != nil {
		return err
	}
	for _, key := range []string{"recording_id", "event_id", "business_id"} {
		if err := optionalString(args, key, 1, 128); err != nil {
			return err
		}
	}
	dateFrom, hasDateFrom, err := optionalRFC3339(args, "date_from")
	if err != nil {
		return err
	}
	dateTo, hasDateTo, err := optionalRFC3339(args, "date_to")
	if err != nil {
		return err
	}
	if hasDateFrom && hasDateTo && !dateFrom.Before(dateTo) {
		return fmt.Errorf("date_from must be before date_to")
	}
	if err := optionalInt(args, "limit", 1, 50); err != nil {
		return err
	}
	if err := optionalString(args, "cursor", 1, 512); err != nil {
		return err
	}
	return rejectUnknown(args, "platform", "request_id", "task_type", "state", "recording_id", "event_id", "business_id", "date_from", "date_to", "limit", "cursor")
}

func validateModelIOFieldArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	lookup := modelio.Lookup{}
	if rawMemory, ok := args["memory_id"]; ok {
		if err := json.Unmarshal(rawMemory, &lookup.MemoryID); err != nil {
			return fmt.Errorf("memory_id must be a string")
		}
	}
	if rawRequest, ok := args["request_id"]; ok {
		if err := json.Unmarshal(rawRequest, &lookup.RequestID); err != nil {
			return fmt.Errorf("request_id must be a string")
		}
	}
	if rawPlatform, ok := args["platform"]; ok {
		if err := json.Unmarshal(rawPlatform, &lookup.Platform); err != nil {
			return fmt.Errorf("platform must be a string")
		}
	}
	if err := modelio.ValidateLookup(lookup); err != nil {
		return err
	}
	if err := optionalString(args, "memory_id", 1, 160); err != nil {
		return err
	}
	if err := optionalString(args, "request_id", 1, 160); err != nil {
		return err
	}
	if err := optionalPlatformValue(args, "platform"); err != nil {
		return err
	}
	if err := optionalString(args, "out", 1, 4096); err != nil {
		return err
	}
	if err := optionalBool(args, "force"); err != nil {
		return err
	}
	return rejectUnknown(args, "memory_id", "request_id", "platform", "out", "force")
}

func optionalModelTaskType(args map[string]json.RawMessage, key string) error {
	value, ok, err := optionalStringValue(args, key, 1, 64)
	if err != nil || !ok {
		return err
	}
	switch value {
	case "transcript_correction", "meeting_summary", "daily_summary", "daily_digest",
		"event_planning", "event_summary", "summary_template_draft", "legacy_classification":
		return nil
	default:
		return fmt.Errorf("%s is invalid", key)
	}
}

func optionalModelTraceState(args map[string]json.RawMessage, key string) error {
	value, ok, err := optionalStringValue(args, key, 1, 64)
	if err != nil || !ok {
		return err
	}
	switch value {
	case "created", "reserved", "executing", "validating", "completed", "provider_failed",
		"reconciliation_required", "response_cache_expired", "trace_failed":
		return nil
	default:
		return fmt.Errorf("%s is invalid", key)
	}
}

func optionalRFC3339(args map[string]json.RawMessage, key string) (time.Time, bool, error) {
	value, ok, err := optionalStringValue(args, key, 1, 64)
	if err != nil || !ok {
		return time.Time{}, ok, err
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, true, fmt.Errorf("%s must be RFC3339", key)
	}
	return parsed, true, nil
}

func optionalStringValue(args map[string]json.RawMessage, key string, minLength int, maxLength int) (string, bool, error) {
	raw, ok := args[key]
	if !ok {
		return "", false, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, fmt.Errorf("%s must be a string", key)
	}
	if len(value) < minLength || len(value) > maxLength {
		return "", true, fmt.Errorf("%s length is out of bounds", key)
	}
	return value, true, nil
}
