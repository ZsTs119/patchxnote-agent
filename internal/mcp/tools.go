package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/api"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/cache"
)

const maxToolOutputBytes = 16 * 1024

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	Annotations map[string]any `json:"annotations,omitempty"`
	handler     ToolHandler
	validator   func(json.RawMessage) error
}

type ToolHandler func(ctx context.Context, arguments json.RawMessage) (CallToolResult, error)

type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func defaultTools(server *Server) []Tool {
	return []Tool{
		{
			Name:        "patchnote_get_current_user",
			Description: "Read the current PatchXNote account projection for the logged-in Agent session.",
			InputSchema: objectSchema(nil, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleGetCurrentUser,
			validator:   validateNoArgs,
		},
		{
			Name:        "patchnote_list_recorder_cards",
			Description: "List recorder cards bound to the current PatchXNote account with masked identifiers only.",
			InputSchema: objectSchema(nil, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleListRecorderCards,
			validator:   validateNoArgs,
		},
		{
			Name:        "patchnote_get_quota_summary",
			Description: "Read the current PatchXNote quota summary for the logged-in account.",
			InputSchema: objectSchema(nil, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleGetQuotaSummary,
			validator:   validateNoArgs,
		},
		{
			Name:        "patchnote_get_model_usage_summary",
			Description: "Read the current-month PatchXNote model usage summary.",
			InputSchema: objectSchema(nil, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleGetModelUsageSummary,
			validator:   validateNoArgs,
		},
		{
			Name:        "patchnote_list_memories",
			Description: "List safe metadata for stored structured results for one selected platform.",
			InputSchema: objectSchema(map[string]any{
				"platform": platformProperty(),
				"limit":    integerProperty(1, 50),
				"cursor":   stringProperty(1, 512),
			}, []string{"platform"}),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleListMemories,
			validator:   validateListMemoriesArgs,
		},
		{
			Name:        "patchnote_search_memories",
			Description: "Search the local authorized PatchXNote memory metadata cache for one selected platform.",
			InputSchema: objectSchema(map[string]any{
				"platform": platformProperty(),
				"query":    stringProperty(1, 128),
				"limit":    integerProperty(1, 50),
			}, []string{"platform", "query"}),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleSearchMemories,
			validator:   validateSearchMemoriesArgs,
		},
		{
			Name:        "patchnote_get_memory",
			Description: "Read safe metadata for one PatchXNote structured result.",
			InputSchema: objectSchema(map[string]any{
				"platform":  platformProperty(),
				"memory_id": stringProperty(1, 160),
			}, []string{"platform", "memory_id"}),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleGetMemory,
			validator:   validateGetMemoryArgs,
		},
	}
}

func (s *Server) handleGetCurrentUser(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	token, err := s.accessToken(ctx, "agent:account.read")
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	user, err := s.api.CurrentUser(ctx, token)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(user)
}

func (s *Server) handleListRecorderCards(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	token, err := s.accessToken(ctx, "agent:hardware.read")
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	cards, err := s.api.ListRecorderCards(ctx, token)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(cards)
}

func (s *Server) handleGetQuotaSummary(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	token, err := s.accessToken(ctx, "agent:quota.read")
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	quota, err := s.api.GetQuotaSummary(ctx, token)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(quota)
}

func (s *Server) handleGetModelUsageSummary(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	token, err := s.accessToken(ctx, "agent:model_usage.read")
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	usage, err := s.api.GetModelUsageSummary(ctx, token)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(usage)
}

func (s *Server) handleListMemories(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	params, err := decodeListMemoriesArgs(arguments)
	if err != nil {
		return CallToolResult{}, err
	}
	token, err := s.accessToken(ctx, "agent:content.read:"+params.Platform)
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	page, err := s.api.ListMemories(ctx, token, params)
	if err != nil {
		return CallToolResult{}, err
	}
	if err := s.memoryCache.UpsertMemories(ctx, apiMemoriesToCache(page.Items)); err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(page)
}

func (s *Server) handleSearchMemories(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	params, err := decodeSearchMemoriesArgs(arguments)
	if err != nil {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", err.Error())
	}
	if _, err := s.accessToken(ctx, "agent:content.read:"+params.Platform); err != nil {
		return CallToolResult{}, err
	}
	result, err := cache.SearchMemories(ctx, s.memoryCache, params)
	if err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(result)
}

func (s *Server) handleGetMemory(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	params, err := decodeGetMemoryArgs(arguments)
	if err != nil {
		return CallToolResult{}, err
	}
	token, err := s.accessToken(ctx, "agent:content.read:"+params.Platform)
	if err != nil {
		return CallToolResult{}, err
	}
	if s.api == nil {
		return CallToolResult{}, rpcErr(codeToolError, "api_unavailable", "PatchXNote API client is not configured")
	}
	memory, err := s.api.GetMemory(ctx, token, params.Platform, params.MemoryID)
	if err != nil {
		return CallToolResult{}, err
	}
	if err := s.memoryCache.UpsertMemories(ctx, apiMemoriesToCache([]api.AgentMemory{memory})); err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(memory)
}

func jsonResult(value any) (CallToolResult, error) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return CallToolResult{}, err
	}
	if len(body) > maxToolOutputBytes {
		return CallToolResult{}, rpcErr(codeToolError, "output_too_large", "tool output exceeded size limit")
	}
	return CallToolResult{
		Content: []ToolContent{
			{Type: "text", Text: string(body)},
		},
	}, nil
}

type getMemoryArgs struct {
	Platform string `json:"platform"`
	MemoryID string `json:"memory_id"`
}

type searchMemoryArgs struct {
	Platform string `json:"platform"`
	Query    string `json:"query"`
	Limit    int    `json:"limit,omitempty"`
}

func decodeListMemoriesArgs(raw json.RawMessage) (api.ListMemoriesParams, error) {
	var args struct {
		Platform string `json:"platform"`
		Limit    int    `json:"limit,omitempty"`
		Cursor   string `json:"cursor,omitempty"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return api.ListMemoriesParams{}, err
	}
	if args.Limit == 0 {
		args.Limit = 20
	}
	return api.ListMemoriesParams{
		Platform: args.Platform,
		Limit:    args.Limit,
		Cursor:   args.Cursor,
	}, nil
}

func decodeSearchMemoriesArgs(raw json.RawMessage) (cache.MemorySearchParams, error) {
	var args searchMemoryArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return cache.MemorySearchParams{}, err
	}
	if args.Limit == 0 {
		args.Limit = 20
	}
	return cache.MemorySearchParams{
		Platform: args.Platform,
		Query:    args.Query,
		Limit:    args.Limit,
	}, nil
}

func decodeGetMemoryArgs(raw json.RawMessage) (getMemoryArgs, error) {
	var args getMemoryArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return getMemoryArgs{}, err
	}
	return args, nil
}

func apiMemoriesToCache(memories []api.AgentMemory) []cache.Memory {
	result := make([]cache.Memory, 0, len(memories))
	for _, memory := range memories {
		result = append(result, cache.Memory{
			ID:                 memory.ID,
			Platform:           memory.Platform,
			ObjectType:         memory.ObjectType,
			ClientObjectID:     memory.ClientObjectID,
			RevisionID:         memory.RevisionID,
			SchemaID:           memory.SchemaID,
			SourceAvailability: memory.SourceAvailability,
			CreatedAt:          memory.CreatedAt,
			UpdatedAt:          memory.UpdatedAt,
		})
	}
	return result
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	if properties == nil {
		properties = map[string]any{}
	}
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func platformProperty() map[string]any {
	return map[string]any{
		"type": "string",
		"enum": []string{"mobile", "desktop"},
	}
}

func stringProperty(minLength int, maxLength int) map[string]any {
	return map[string]any{
		"type":      "string",
		"minLength": minLength,
		"maxLength": maxLength,
	}
}

func integerProperty(minimum int, maximum int) map[string]any {
	return map[string]any{
		"type":    "integer",
		"minimum": minimum,
		"maximum": maximum,
	}
}

func readOnlyAnnotations() map[string]any {
	return map[string]any{"readOnlyHint": true}
}

func validateNoArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if len(args) != 0 {
		return fmt.Errorf("arguments must be empty")
	}
	return nil
}

func validateListMemoriesArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := requirePlatform(args); err != nil {
		return err
	}
	if err := optionalInt(args, "limit", 1, 50); err != nil {
		return err
	}
	if err := optionalString(args, "cursor", 1, 512); err != nil {
		return err
	}
	return rejectUnknown(args, "platform", "limit", "cursor")
}

func validateSearchMemoriesArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := requirePlatform(args); err != nil {
		return err
	}
	if err := requireString(args, "query", 1, 128); err != nil {
		return err
	}
	if err := optionalInt(args, "limit", 1, 50); err != nil {
		return err
	}
	return rejectUnknown(args, "platform", "query", "limit")
}

func validateGetMemoryArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := requirePlatform(args); err != nil {
		return err
	}
	if err := requireString(args, "memory_id", 1, 160); err != nil {
		return err
	}
	return rejectUnknown(args, "platform", "memory_id")
}

func decodeArgs(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]json.RawMessage{}, nil
	}
	var args map[string]json.RawMessage
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("arguments must be a JSON object")
	}
	if args == nil {
		return map[string]json.RawMessage{}, nil
	}
	return args, nil
}

func requirePlatform(args map[string]json.RawMessage) error {
	var platform string
	if err := requireInto(args, "platform", &platform); err != nil {
		return err
	}
	if platform != "mobile" && platform != "desktop" {
		return fmt.Errorf("platform must be mobile or desktop")
	}
	return nil
}

func requireString(args map[string]json.RawMessage, key string, minLength int, maxLength int) error {
	var value string
	if err := requireInto(args, key, &value); err != nil {
		return err
	}
	if len(value) < minLength || len(value) > maxLength {
		return fmt.Errorf("%s length is out of bounds", key)
	}
	return nil
}

func optionalString(args map[string]json.RawMessage, key string, minLength int, maxLength int) error {
	raw, ok := args[key]
	if !ok {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s must be a string", key)
	}
	if len(value) < minLength || len(value) > maxLength {
		return fmt.Errorf("%s length is out of bounds", key)
	}
	return nil
}

func optionalInt(args map[string]json.RawMessage, key string, minimum int, maximum int) error {
	raw, ok := args[key]
	if !ok {
		return nil
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s must be an integer", key)
	}
	if value < minimum || value > maximum {
		return fmt.Errorf("%s is out of bounds", key)
	}
	return nil
}

func requireInto(args map[string]json.RawMessage, key string, target any) error {
	raw, ok := args[key]
	if !ok {
		return fmt.Errorf("%s is required", key)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("%s has invalid type", key)
	}
	return nil
}

func rejectUnknown(args map[string]json.RawMessage, allowed ...string) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range args {
		if _, ok := allowedSet[key]; !ok {
			return fmt.Errorf("unknown argument %s", key)
		}
	}
	return nil
}
