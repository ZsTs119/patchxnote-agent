package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/cache"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/webhook"
)

const (
	protocolVersion      = "2025-06-18"
	defaultServerVersion = "0.0.0-dev"
)

type Authenticator interface {
	Status(ctx context.Context) (auth.Status, error)
}

type CredentialProvider interface {
	Credential(ctx context.Context) (keychain.Credential, bool, error)
}

type RefreshingCredentialProvider interface {
	CredentialProvider
	RefreshNow(ctx context.Context) (keychain.Credential, bool, error)
}

type AgentAPI interface {
	CurrentUser(ctx context.Context, accessToken string) (api.CurrentAccount, error)
	ListRecorderCards(ctx context.Context, accessToken string) (api.AgentRecorderCardPage, error)
	GetQuotaSummary(ctx context.Context, accessToken string) (api.AgentQuotaSummary, error)
	GetModelUsageSummary(ctx context.Context, accessToken string) (api.AgentModelUsageSummary, error)
	ListMemories(ctx context.Context, accessToken string, params api.ListMemoriesParams) (api.AgentMemoryPage, error)
	GetMemory(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentMemory, error)
	GetMemoryDeliveryDocument(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentDeliveryDocument, error)
	GetMemoryModelIO(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentModelIOExport, error)
	GetModelRunIOTrace(ctx context.Context, accessToken string, platform string, requestID string) (api.AgentModelIOExport, error)
	ListModelIOTraces(ctx context.Context, accessToken string, params api.ListModelIOTracesParams) (api.AgentModelIOTracePage, error)
}

type Options struct {
	Authenticator Authenticator
	Credentials   CredentialProvider
	API           AgentAPI
	MemoryCache   *cache.MemoryIndex
	Config        config.Config
	Secrets       keychain.SecretStore
	Tools         []Tool
	Version       string
}

type Server struct {
	authenticator Authenticator
	credentials   CredentialProvider
	api           AgentAPI
	memoryCache   *cache.MemoryIndex
	config        config.Config
	secrets       keychain.SecretStore
	tools         map[string]Tool
	toolList      []Tool
	version       string
}

func NewServer(options Options) *Server {
	serverVersion := options.Version
	if serverVersion == "" {
		serverVersion = defaultServerVersion
	}
	server := &Server{
		authenticator: options.Authenticator,
		credentials:   options.Credentials,
		api:           options.API,
		memoryCache:   options.MemoryCache,
		config:        options.Config,
		secrets:       options.Secrets,
		tools:         make(map[string]Tool),
		version:       serverVersion,
	}
	if server.credentials == nil {
		if provider, ok := options.Authenticator.(CredentialProvider); ok {
			server.credentials = provider
		}
	}
	if server.memoryCache == nil {
		server.memoryCache = cache.NewMemoryIndex()
	}
	if server.config.Profile == "" {
		server.config.Profile = "default"
	}
	if server.secrets == nil {
		server.secrets = keychain.UnavailableStore{}
	}
	tools := options.Tools
	if len(tools) == 0 {
		tools = defaultTools(server)
	}
	for _, tool := range tools {
		server.tools[tool.Name] = tool
		server.toolList = append(server.toolList, tool)
	}
	return server
}

func (s *Server) Serve(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	encoder := json.NewEncoder(stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		response, shouldRespond := s.handleLine(ctx, line)
		if !shouldRespond {
			continue
		}
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("write mcp response: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read mcp request: %w", err)
	}
	return nil
}

func (s *Server) handleLine(ctx context.Context, line []byte) (rpcResponse, bool) {
	var request rpcRequest
	if err := json.Unmarshal(line, &request); err != nil {
		return errorResponse(nil, rpcErr(codeParseError, "parse_error", "invalid JSON-RPC message")), true
	}
	if len(request.ID) == 0 && isNotification(request.Method) {
		return rpcResponse{}, false
	}
	if request.JSONRPC != "2.0" || request.Method == "" {
		return errorResponse(request.ID, rpcErr(codeInvalidRequest, "invalid_request", "invalid JSON-RPC request")), true
	}

	switch request.Method {
	case "initialize":
		return resultResponse(request.ID, initializeResult{
			ProtocolVersion: protocolVersion,
			ServerInfo: serverInfo{
				Name:    "patchxnote-agent",
				Version: s.version,
			},
			Capabilities: serverCapabilities{
				Tools: map[string]any{},
			},
		}), true
	case "ping":
		return resultResponse(request.ID, map[string]any{}), true
	case "tools/list":
		return resultResponse(request.ID, toolsListResult{Tools: s.publicTools()}), true
	case "tools/call":
		result, err := s.callTool(ctx, request.Params)
		if err != nil {
			return errorResponse(request.ID, mapToolError(err)), true
		}
		return resultResponse(request.ID, result), true
	default:
		return errorResponse(request.ID, rpcErr(codeMethodNotFound, "method_not_found", "method not found")), true
	}
}

func (s *Server) publicTools() []Tool {
	tools := make([]Tool, 0, len(s.toolList))
	for _, tool := range s.toolList {
		tool.handler = nil
		tool.validator = nil
		tools = append(tools, tool)
	}
	return tools
}

func (s *Server) callTool(ctx context.Context, raw json.RawMessage) (CallToolResult, error) {
	var params callToolParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", "tools/call params are invalid")
	}
	if params.Name == "" {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", "tool name is required")
	}
	tool, ok := s.tools[params.Name]
	if !ok {
		return CallToolResult{}, rpcErr(codeInvalidParams, "tool_not_found", "tool not found")
	}
	if tool.validator != nil {
		if err := tool.validator(params.Arguments); err != nil {
			return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", err.Error())
		}
	}
	if tool.handler == nil {
		return CallToolResult{}, rpcErr(codeToolError, "tool_unavailable", "tool handler is unavailable")
	}
	return tool.handler(ctx, params.Arguments)
}

func (s *Server) authRequiredHandler(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	if _, err := s.accessToken(ctx); err != nil {
		return CallToolResult{}, err
	}
	return CallToolResult{}, rpcErr(codeToolError, "tool_not_implemented", "tool implementation is pending")
}

func (s *Server) accessToken(ctx context.Context, requiredScopes ...string) (string, error) {
	if s.credentials == nil {
		return "", authRequiredError()
	}
	credential, ok, err := s.credentials.Credential(ctx)
	if err != nil {
		return "", err
	}
	if !ok || credential.AccessToken == "" {
		return "", authRequiredError()
	}
	for _, scope := range requiredScopes {
		if !hasScope(credential.Scopes, scope) {
			return "", rpcErr(codeToolError, "permission_denied", "PatchXNote Agent does not have permission for this tool")
		}
	}
	return credential.AccessToken, nil
}

func hasScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

func mapToolError(err error) *Error {
	var mcpErr *Error
	if errors.As(err, &mcpErr) {
		return mcpErr
	}
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		return mapAPIError(apiErr)
	}
	if errors.Is(err, webhook.ErrTargetNotFound) {
		return rpcErr(codeToolError, "webhook_target_not_found", "Webhook target was not found")
	}
	if errors.Is(err, webhook.ErrWebhookSecretMissing) {
		return rpcErr(codeToolError, "webhook_secret_missing", "Webhook URL is missing; configure the target again")
	}
	if isUserFacingToolError(err) {
		return rpcErr(codeToolError, "tool_error", err.Error())
	}
	return rpcErr(codeInternalError, "internal_error", "internal MCP server error")
}

func mapAPIError(err *api.Error) *Error {
	switch err.StatusCode {
	case 401:
		return rpcErr(codeAuthRequired, "auth_required", "PatchXNote Agent login is required")
	case 403:
		return rpcErr(codeToolError, "permission_denied", "PatchXNote Agent does not have permission for this tool")
	case 404:
		return rpcErr(codeToolError, "not_found", "PatchXNote resource was not found")
	case 429:
		return rpcErr(codeToolError, "rate_limited", "PatchXNote API rate limit reached")
	default:
		if err.Code != "" {
			return rpcErr(codeToolError, err.Code, "PatchXNote API request failed")
		}
		return rpcErr(codeToolError, "api_error", "PatchXNote API request failed")
	}
}

func isNotification(method string) bool {
	return method == "notifications/initialized" || method == "notifications/cancelled"
}
