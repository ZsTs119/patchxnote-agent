package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/renderdoc"
	"github.com/ZsTs119/patchxnote-agent/internal/webhook"
)

const (
	maxDirectMarkdownBytes = renderdoc.MaxRenderedMarkdownBytes
	minWebhookTimeout      = time.Second
	maxWebhookTimeout      = 120 * time.Second
)

type webhookTargetResult struct {
	Alias        string `json:"alias"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	MaskedURL    string `json:"masked_url"`
	Template     string `json:"template,omitempty"`
	SecretStatus string `json:"secret_status"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type webhookRenderResult struct {
	Title     string                   `json:"title"`
	Markdown  string                   `json:"markdown,omitempty"`
	Memory    *api.AgentDeliveryMemory `json:"memory,omitempty"`
	Trace     api.AgentDeliveryTrace   `json:"trace"`
	Template  string                   `json:"template"`
	DraftPath string                   `json:"draft_path,omitempty"`
	Files     map[string]string        `json:"files,omitempty"`
}

type webhookExportResult struct {
	Source   string                      `json:"source"`
	Memory   *api.AgentDeliveryMemory    `json:"memory,omitempty"`
	Trace    api.AgentDeliveryTrace      `json:"trace"`
	Status   api.AgentModelIOFieldStatus `json:"field_status"`
	Output   string                      `json:"out"`
	Exported bool                        `json:"exported"`
}

func defaultWebhookTools(server *Server) []Tool {
	return []Tool{
		{
			Name:        "patchxnote_list_webhook_targets",
			Description: "List local webhook target aliases and masked metadata for the active PatchXNote Agent profile.",
			InputSchema: objectSchema(map[string]any{
				"include_disabled": booleanProperty(),
			}, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleListWebhookTargets,
			validator:   validateListWebhookTargetsArgs,
		},
		{
			Name:        "patchxnote_configure_webhook_target",
			Description: "Create or update a local webhook target. Webhook URL and signing secret are write-only inputs and are never returned.",
			InputSchema: objectSchema(map[string]any{
				"alias":                stringProperty(1, 64),
				"type":                 targetTypeProperty(),
				"webhook_url":          stringProperty(1, 4096),
				"signing_secret":       stringProperty(1, 4096),
				"clear_signing_secret": booleanProperty(),
				"enabled":              booleanProperty(),
				"template":             stringProperty(1, 512),
			}, []string{"alias"}),
			Annotations: localWriteAnnotations(),
			handler:     server.handleConfigureWebhookTarget,
			validator:   validateConfigureWebhookTargetArgs,
		},
		{
			Name:        "patchxnote_remove_webhook_target",
			Description: "Remove a local webhook target alias and best-effort clean up its stored webhook secrets.",
			InputSchema: objectSchema(map[string]any{
				"alias": stringProperty(1, 64),
			}, []string{"alias"}),
			Annotations: localDeleteAnnotations(),
			handler:     server.handleRemoveWebhookTarget,
			validator:   validateRemoveWebhookTargetArgs,
		},
		{
			Name:        "patchxnote_list_webhook_templates",
			Description: "List built-in webhook Markdown templates. Custom templates are passed as local file paths to render or send tools.",
			InputSchema: objectSchema(nil, nil),
			Annotations: readOnlyAnnotations(),
			handler:     server.handleListWebhookTemplates,
			validator:   validateNoArgs,
		},
		{
			Name:        "patchxnote_render_webhook_message",
			Description: "Render a PatchXNote memory delivery document into Markdown and optionally save a local draft directory.",
			InputSchema: objectSchema(map[string]any{
				"memory_id":        stringProperty(1, 160),
				"platform":         platformProperty(),
				"template":         stringProperty(1, 512),
				"title":            stringProperty(1, 256),
				"save_draft":       booleanProperty(),
				"out":              stringProperty(1, 4096),
				"include_model_io": booleanProperty(),
				"force":            booleanProperty(),
			}, []string{"memory_id"}),
			Annotations: localWriteAnnotations(),
			handler:     server.handleRenderWebhookMessage,
			validator:   validateRenderWebhookMessageArgs,
		},
		{
			Name:        "patchxnote_export_model_io",
			Description: "Export explicit PatchXNote Agent model IO JSON to a user-chosen local file and return only a summary.",
			InputSchema: objectSchema(map[string]any{
				"memory_id":  stringProperty(1, 160),
				"request_id": stringProperty(1, 160),
				"platform":   platformProperty(),
				"out":        stringProperty(1, 4096),
				"force":      booleanProperty(),
			}, []string{"out"}),
			Annotations: localWriteAnnotations(),
			handler:     server.handleExportWebhookModelIO,
			validator:   validateExportWebhookModelIOArgs,
		},
		{
			Name:        "patchxnote_send_webhook",
			Description: "Send a webhook message to one or more configured target aliases. This performs external network requests.",
			InputSchema: objectSchema(map[string]any{
				"target_aliases":   arrayProperty(stringProperty(1, 64), 1, 20),
				"title":            stringProperty(1, 256),
				"markdown":         stringProperty(1, maxDirectMarkdownBytes),
				"draft_dir":        stringProperty(1, 4096),
				"memory_id":        stringProperty(1, 160),
				"platform":         platformProperty(),
				"template":         stringProperty(1, 512),
				"test_message":     booleanProperty(),
				"timeout_seconds":  integerProperty(1, 120),
				"save_draft":       booleanProperty(),
				"out":              stringProperty(1, 4096),
				"include_model_io": booleanProperty(),
				"force":            booleanProperty(),
			}, []string{"target_aliases"}),
			Annotations: externalSendAnnotations(),
			handler:     server.handleSendWebhook,
			validator:   validateSendWebhookArgs,
		},
	}
}

func (s *Server) handleListWebhookTargets(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	args := struct {
		IncludeDisabled *bool `json:"include_disabled,omitempty"`
	}{}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	includeDisabled := true
	if args.IncludeDisabled != nil {
		includeDisabled = *args.IncludeDisabled
	}
	registry, err := s.webhookRegistry()
	if err != nil {
		return CallToolResult{}, err
	}
	secrets := webhook.NewSecretStore(s.secrets, s.config.Profile)
	targets := registry.List()
	result := make([]webhookTargetResult, 0, len(targets))
	for _, target := range targets {
		if !includeDisabled && !target.Enabled {
			continue
		}
		result = append(result, s.webhookTargetOutput(ctx, secrets, target))
	}
	return jsonResult(result)
}

func (s *Server) handleConfigureWebhookTarget(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	var args struct {
		Alias              string `json:"alias"`
		Type               string `json:"type,omitempty"`
		WebhookURL         string `json:"webhook_url,omitempty"`
		SigningSecret      string `json:"signing_secret,omitempty"`
		ClearSigningSecret bool   `json:"clear_signing_secret,omitempty"`
		Enabled            *bool  `json:"enabled,omitempty"`
		Template           string `json:"template,omitempty"`
	}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	rawArgs, err := decodeArgs(arguments)
	if err != nil {
		return CallToolResult{}, err
	}
	alias, err := webhook.ValidateAlias(args.Alias)
	if err != nil {
		return CallToolResult{}, err
	}
	registry, err := s.webhookRegistry()
	if err != nil {
		return CallToolResult{}, err
	}
	secrets := webhook.NewSecretStore(s.secrets, s.config.Profile)
	existing, existingErr := registry.Get(alias)
	exists := existingErr == nil
	if existingErr != nil && !errors.Is(existingErr, webhook.ErrTargetNotFound) {
		return CallToolResult{}, existingErr
	}

	if !exists && strings.TrimSpace(args.WebhookURL) == "" {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", "webhook_url is required when creating a target")
	}
	targetType := existing.Type
	if strings.TrimSpace(args.Type) != "" {
		got, err := webhook.ValidateType(args.Type)
		if err != nil {
			return CallToolResult{}, err
		}
		targetType = got
	} else if !exists {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", "type is required when creating a target")
	}
	if exists && targetType != existing.Type && strings.TrimSpace(args.WebhookURL) == "" {
		return CallToolResult{}, rpcErr(codeInvalidParams, "invalid_params", "webhook_url is required when changing target type")
	}

	cleanupSecretsOnCreateFailure := !exists
	maskedURL := existing.MaskedURL
	if strings.TrimSpace(args.WebhookURL) != "" {
		validated, err := webhook.ValidateWebhookURL(args.WebhookURL)
		if err != nil {
			return CallToolResult{}, err
		}
		if err := secrets.PutURL(ctx, alias, validated); err != nil {
			return CallToolResult{}, err
		}
		maskedURL = webhook.MaskWebhookURL(validated)
	}
	if args.SigningSecret != "" {
		if err := secrets.PutSigningSecret(ctx, alias, args.SigningSecret); err != nil {
			return CallToolResult{}, err
		}
	} else if args.ClearSigningSecret {
		if err := secrets.DeleteSigningSecret(ctx, alias); err != nil && !isSecretMissing(err) {
			return CallToolResult{}, err
		}
	}
	templateName := existing.Template
	if _, ok := rawArgs["template"]; ok {
		templateName = args.Template
	}
	target, err := registry.Set(alias, targetType, maskedURL, templateName)
	if err != nil {
		if cleanupSecretsOnCreateFailure {
			_ = secrets.DeleteTarget(ctx, alias)
		}
		return CallToolResult{}, err
	}
	if args.Enabled != nil {
		target, err = registry.SetEnabled(alias, *args.Enabled)
		if err != nil {
			if cleanupSecretsOnCreateFailure {
				_ = secrets.DeleteTarget(ctx, alias)
			}
			return CallToolResult{}, err
		}
	}
	return jsonResult(s.webhookTargetOutput(ctx, secrets, target))
}

func (s *Server) handleRemoveWebhookTarget(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	var args struct {
		Alias string `json:"alias"`
	}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	registry, err := s.webhookRegistry()
	if err != nil {
		return CallToolResult{}, err
	}
	target, err := registry.Remove(args.Alias)
	if err != nil {
		return CallToolResult{}, err
	}
	secrets := webhook.NewSecretStore(s.secrets, s.config.Profile)
	cleanup := "ok"
	if err := secrets.DeleteTarget(ctx, target.Alias); err != nil && !isSecretMissing(err) {
		cleanup = err.Error()
	}
	return jsonResult(map[string]any{
		"removed":        true,
		"alias":          target.Alias,
		"type":           string(target.Type),
		"masked_url":     target.MaskedURL,
		"cleanup_status": cleanup,
	})
}

func (s *Server) handleListWebhookTemplates(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	return jsonResult(renderdoc.BuiltInTemplates())
}

func (s *Server) handleRenderWebhookMessage(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	var args struct {
		MemoryID       string `json:"memory_id"`
		Platform       string `json:"platform,omitempty"`
		Template       string `json:"template,omitempty"`
		Title          string `json:"title,omitempty"`
		SaveDraft      bool   `json:"save_draft,omitempty"`
		Out            string `json:"out,omitempty"`
		IncludeModelIO bool   `json:"include_model_io,omitempty"`
		Force          bool   `json:"force,omitempty"`
	}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	delivery, rendered, modelIO, err := s.renderMemoryWebhook(ctx, args.MemoryID, args.Platform, args.Template, args.IncludeModelIO)
	if err != nil {
		return CallToolResult{}, err
	}
	title := strings.TrimSpace(args.Title)
	if title == "" {
		title = delivery.Title
	}
	result := webhookRenderResult{
		Title:    title,
		Markdown: rendered,
		Memory:   delivery.Memory,
		Trace:    delivery.Trace,
		Template: webhook.TemplateNameOrDefault(args.Template),
	}
	if args.SaveDraft {
		out, err := absPath(args.Out)
		if err != nil {
			return CallToolResult{}, err
		}
		if err := webhook.WriteDraftDirectory(out, delivery, rendered, webhook.TemplateNameOrDefault(args.Template), modelIO, args.Force); err != nil {
			return CallToolResult{}, err
		}
		result.DraftPath = out
		result.Files = draftFiles(out, modelIO != nil)
		if len(rendered) > maxToolOutputBytes {
			result.Markdown = ""
		}
	} else if len(rendered) > maxToolOutputBytes {
		return CallToolResult{}, rpcErr(codeToolError, "output_too_large", "rendered Markdown exceeds MCP output limit; rerun with save_draft=true and out")
	}
	return jsonResult(result)
}

func (s *Server) handleExportWebhookModelIO(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	var args struct {
		MemoryID  string `json:"memory_id,omitempty"`
		RequestID string `json:"request_id,omitempty"`
		Platform  string `json:"platform,omitempty"`
		Out       string `json:"out"`
		Force     bool   `json:"force,omitempty"`
	}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	var export api.AgentModelIOExport
	var err error
	if args.MemoryID != "" {
		export, err = s.fetchMemoryModelIO(ctx, args.Platform, args.MemoryID)
	} else {
		export, err = s.fetchModelRunIOTrace(ctx, args.Platform, args.RequestID)
	}
	if err != nil {
		return CallToolResult{}, err
	}
	out, err := absPath(args.Out)
	if err != nil {
		return CallToolResult{}, err
	}
	body, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return CallToolResult{}, fmt.Errorf("encode model IO export: %w", err)
	}
	body = append(body, '\n')
	if err := webhook.AtomicWriteFile(out, body, 0o600, args.Force); err != nil {
		return CallToolResult{}, err
	}
	return jsonResult(webhookExportResult{
		Source:   export.Source,
		Memory:   export.Memory,
		Trace:    export.Trace,
		Status:   export.FieldStatus,
		Output:   out,
		Exported: true,
	})
}

func (s *Server) handleSendWebhook(ctx context.Context, arguments json.RawMessage) (CallToolResult, error) {
	var args struct {
		TargetAliases  []string `json:"target_aliases"`
		Title          string   `json:"title,omitempty"`
		Markdown       string   `json:"markdown,omitempty"`
		DraftDir       string   `json:"draft_dir,omitempty"`
		MemoryID       string   `json:"memory_id,omitempty"`
		Platform       string   `json:"platform,omitempty"`
		Template       string   `json:"template,omitempty"`
		TestMessage    bool     `json:"test_message,omitempty"`
		TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
		SaveDraft      bool     `json:"save_draft,omitempty"`
		Out            string   `json:"out,omitempty"`
		IncludeModelIO bool     `json:"include_model_io,omitempty"`
		Force          bool     `json:"force,omitempty"`
	}
	if err := json.Unmarshal(argumentsOrEmpty(arguments), &args); err != nil {
		return CallToolResult{}, err
	}
	if args.MemoryID != "" && strings.TrimSpace(args.Template) == "" && len(args.TargetAliases) == 1 {
		if templateName, err := s.targetTemplate(args.TargetAliases[0]); err == nil && strings.TrimSpace(templateName) != "" {
			args.Template = templateName
		}
	}
	message, err := s.webhookMessageFromSendArgs(ctx, args.Title, args.Markdown, args.DraftDir, args.MemoryID, args.Platform, args.Template, args.TestMessage, args.SaveDraft, args.Out, args.IncludeModelIO, args.Force)
	if err != nil {
		return CallToolResult{}, err
	}
	resolved, err := s.resolveWebhookTargets(ctx, args.TargetAliases)
	if err != nil {
		return CallToolResult{}, err
	}
	timeout := webhook.DefaultSendTimeout
	if args.TimeoutSeconds > 0 {
		timeout = time.Duration(args.TimeoutSeconds) * time.Second
	}
	results, sendErr := webhook.NewSender(nil).Send(ctx, resolved, message, webhook.SendOptions{Timeout: timeout})
	if sendErr != nil {
		result, err := jsonResult(map[string]any{"success": false, "results": results, "error": sendErr.Error()})
		result.IsError = true
		return result, err
	}
	return jsonResult(map[string]any{"success": true, "results": results})
}

func (s *Server) webhookMessageFromSendArgs(ctx context.Context, title string, markdown string, draftDir string, memoryID string, platform string, templateName string, testMessage bool, saveDraft bool, outDir string, includeModelIO bool, force bool) (webhook.Message, error) {
	switch {
	case testMessage:
		return webhook.Message{
			Title:    "PatchXNote Webhook Test",
			Markdown: "这是一条 PatchXNote Agent MCP 本地测试消息。",
			Metadata: map[string]string{"source": "mcp-test"},
		}, nil
	case markdown != "":
		return webhook.MessageFromMarkdown(markdown, title)
	case draftDir != "":
		return webhook.MessageFromDraft(draftDir)
	case memoryID != "":
		delivery, rendered, modelIO, err := s.renderMemoryWebhook(ctx, memoryID, platform, templateName, includeModelIO)
		if err != nil {
			return webhook.Message{}, err
		}
		if saveDraft {
			out, err := absPath(outDir)
			if err != nil {
				return webhook.Message{}, err
			}
			if err := webhook.WriteDraftDirectory(out, delivery, rendered, webhook.TemplateNameOrDefault(templateName), modelIO, force); err != nil {
				return webhook.Message{}, err
			}
		}
		message := webhook.MessageFromDelivery(delivery, rendered, webhook.TemplateNameOrDefault(templateName))
		if strings.TrimSpace(title) != "" {
			message.Title = strings.TrimSpace(title)
		}
		return message, nil
	default:
		return webhook.Message{}, fmt.Errorf("missing webhook message source")
	}
}

func (s *Server) renderMemoryWebhook(ctx context.Context, memoryID string, platform string, templateName string, includeModelIO bool) (api.AgentDeliveryDocument, string, *api.AgentModelIOExport, error) {
	delivery, err := s.fetchDeliveryDocument(ctx, platform, memoryID)
	if err != nil {
		return api.AgentDeliveryDocument{}, "", nil, err
	}
	doc := renderdoc.FromDeliveryDocument(delivery)
	rendered, err := renderdoc.RenderTemplate(templateName, doc)
	if err != nil {
		return api.AgentDeliveryDocument{}, "", nil, err
	}
	var modelIO *api.AgentModelIOExport
	if includeModelIO {
		got, err := s.fetchMemoryModelIO(ctx, platform, memoryID)
		if err != nil {
			return api.AgentDeliveryDocument{}, "", nil, err
		}
		modelIO = &got
	}
	return delivery, rendered, modelIO, nil
}

func (s *Server) resolveWebhookTargets(ctx context.Context, aliases []string) ([]webhook.ResolvedTarget, error) {
	normalizedAliases := make([]string, 0, len(aliases))
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		normalized, err := webhook.ValidateAlias(alias)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			return nil, fmt.Errorf("duplicate webhook target alias %q", normalized)
		}
		seen[normalized] = struct{}{}
		normalizedAliases = append(normalizedAliases, normalized)
	}
	registry, err := s.webhookRegistry()
	if err != nil {
		return nil, err
	}
	secrets := webhook.NewSecretStore(s.secrets, s.config.Profile)
	resolved := make([]webhook.ResolvedTarget, 0, len(normalizedAliases))
	for _, alias := range normalizedAliases {
		target, err := registry.Get(alias)
		if err != nil {
			return nil, err
		}
		if !target.Enabled {
			return nil, fmt.Errorf("webhook target %q is disabled", target.Alias)
		}
		rawURL, err := secrets.URL(ctx, target.Alias)
		if err != nil {
			return nil, err
		}
		signingSecret, _, err := secrets.SigningSecret(ctx, target.Alias)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, webhook.ResolvedTarget{
			Target:        target,
			URL:           rawURL,
			SigningSecret: signingSecret,
		})
	}
	return resolved, nil
}

func (s *Server) fetchDeliveryDocument(ctx context.Context, platform string, memoryID string) (api.AgentDeliveryDocument, error) {
	var delivery api.AgentDeliveryDocument
	err := s.withAgentAccessToken(ctx, platform, func(accessToken string) error {
		var err error
		delivery, err = s.api.GetMemoryDeliveryDocument(ctx, accessToken, platform, memoryID)
		return err
	})
	return delivery, friendlyAgentMCPError(err)
}

func (s *Server) fetchMemoryModelIO(ctx context.Context, platform string, memoryID string) (api.AgentModelIOExport, error) {
	var export api.AgentModelIOExport
	err := s.withAgentAccessToken(ctx, platform, func(accessToken string) error {
		var err error
		export, err = s.api.GetMemoryModelIO(ctx, accessToken, platform, memoryID)
		return err
	})
	return export, friendlyAgentMCPError(err)
}

func (s *Server) fetchModelRunIOTrace(ctx context.Context, platform string, requestID string) (api.AgentModelIOExport, error) {
	var export api.AgentModelIOExport
	err := s.withAgentAccessToken(ctx, platform, func(accessToken string) error {
		var err error
		export, err = s.api.GetModelRunIOTrace(ctx, accessToken, platform, requestID)
		return err
	})
	return export, friendlyAgentMCPError(err)
}

func (s *Server) withAgentAccessToken(ctx context.Context, platform string, call func(accessToken string) error) error {
	if s.api == nil {
		return fmt.Errorf("PatchXNote API client is not configured")
	}
	requiredScopes := []string{}
	if platform != "" {
		requiredScopes = append(requiredScopes, "agent:content.read:"+platform)
	}
	accessToken, err := s.accessToken(ctx, requiredScopes...)
	if err != nil {
		return err
	}
	err = call(accessToken)
	if isUnauthorizedAPIError(err) {
		if refresher, ok := s.credentials.(RefreshingCredentialProvider); ok {
			refreshed, refreshedOK, refreshErr := refresher.RefreshNow(ctx)
			if refreshErr != nil {
				return refreshErr
			}
			if refreshedOK && refreshed.AccessToken != "" {
				for _, scope := range requiredScopes {
					if !hasScope(refreshed.Scopes, scope) {
						return rpcErr(codeToolError, "permission_denied", "PatchXNote Agent does not have permission for this tool")
					}
				}
				err = call(refreshed.AccessToken)
			}
		}
	}
	return err
}

func (s *Server) webhookTargetOutput(ctx context.Context, secrets *webhook.SecretStore, target webhook.Target) webhookTargetResult {
	secretStatus := "missing"
	if target.Enabled {
		if _, err := secrets.URL(ctx, target.Alias); err == nil {
			secretStatus = "ok"
		}
	} else {
		secretStatus = "disabled"
	}
	return webhookTargetResult{
		Alias:        target.Alias,
		Type:         string(target.Type),
		Enabled:      target.Enabled,
		MaskedURL:    target.MaskedURL,
		Template:     target.Template,
		SecretStatus: secretStatus,
		CreatedAt:    formatToolTime(target.CreatedAt),
		UpdatedAt:    formatToolTime(target.UpdatedAt),
	}
}

func (s *Server) targetTemplate(alias string) (string, error) {
	registry, err := s.webhookRegistry()
	if err != nil {
		return "", err
	}
	target, err := registry.Get(alias)
	if err != nil {
		return "", err
	}
	return target.Template, nil
}

func (s *Server) webhookRegistry() (*webhook.Registry, error) {
	cfg := s.config
	if cfg.Paths.ConfigFile != "" {
		if _, err := os.Stat(cfg.Paths.ConfigFile); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			if cfg.Profile == "" {
				cfg.Profile = "default"
			}
			return webhook.NewRegistry(cfg), nil
		}
		loaded, err := config.Load(config.NewViper(), config.LoadOptions{ConfigFile: cfg.Paths.ConfigFile})
		if err != nil {
			return nil, err
		}
		if cfg.Profile != "" {
			loaded.Profile = cfg.Profile
		}
		cfg = loaded
		s.config.Webhooks = loaded.Webhooks
		s.config.Paths = loaded.Paths
	}
	if cfg.Profile == "" {
		cfg.Profile = "default"
	}
	return webhook.NewRegistry(cfg), nil
}

func argumentsOrEmpty(raw json.RawMessage) []byte {
	if len(raw) == 0 || string(raw) == "null" {
		return []byte("{}")
	}
	return raw
}

func booleanProperty() map[string]any {
	return map[string]any{"type": "boolean"}
}

func arrayProperty(items map[string]any, minItems int, maxItems int) map[string]any {
	return map[string]any{
		"type":     "array",
		"items":    items,
		"minItems": minItems,
		"maxItems": maxItems,
	}
}

func targetTypeProperty() map[string]any {
	return map[string]any{
		"type": "string",
		"enum": []string{"feishu", "dingtalk", "generic"},
	}
}

func localWriteAnnotations() map[string]any {
	return map[string]any{"readOnlyHint": false, "destructiveHint": false}
}

func localDeleteAnnotations() map[string]any {
	return map[string]any{"readOnlyHint": false, "destructiveHint": true}
}

func externalSendAnnotations() map[string]any {
	return map[string]any{"readOnlyHint": false, "destructiveHint": false, "openWorldHint": true}
}

func validateListWebhookTargetsArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := optionalBool(args, "include_disabled"); err != nil {
		return err
	}
	return rejectUnknown(args, "include_disabled")
}

func validateConfigureWebhookTargetArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	var alias string
	if err := requireInto(args, "alias", &alias); err != nil {
		return err
	}
	if _, err := webhook.ValidateAlias(alias); err != nil {
		return err
	}
	if err := optionalTargetType(args, "type"); err != nil {
		return err
	}
	if rawURL, ok := args["webhook_url"]; ok {
		var value string
		if err := json.Unmarshal(rawURL, &value); err != nil {
			return fmt.Errorf("webhook_url must be a string")
		}
		if _, err := webhook.ValidateWebhookURL(value); err != nil {
			return err
		}
	}
	if err := optionalString(args, "signing_secret", 1, 4096); err != nil {
		return err
	}
	if err := optionalBool(args, "clear_signing_secret"); err != nil {
		return err
	}
	if err := optionalBool(args, "enabled"); err != nil {
		return err
	}
	if err := optionalString(args, "template", 1, 512); err != nil {
		return err
	}
	if _, hasSecret := args["signing_secret"]; hasSecret {
		if clear, ok, err := optionalBoolValue(args, "clear_signing_secret"); err != nil {
			return err
		} else if ok && clear {
			return fmt.Errorf("signing_secret conflicts with clear_signing_secret")
		}
	}
	return rejectUnknown(args, "alias", "type", "webhook_url", "signing_secret", "clear_signing_secret", "enabled", "template")
}

func validateRemoveWebhookTargetArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	var alias string
	if err := requireInto(args, "alias", &alias); err != nil {
		return err
	}
	if _, err := webhook.ValidateAlias(alias); err != nil {
		return err
	}
	return rejectUnknown(args, "alias")
}

func validateRenderWebhookMessageArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	if err := requireString(args, "memory_id", 1, 160); err != nil {
		return err
	}
	if err := optionalPlatformValue(args, "platform"); err != nil {
		return err
	}
	if err := optionalString(args, "template", 1, 512); err != nil {
		return err
	}
	if err := optionalString(args, "title", 1, 256); err != nil {
		return err
	}
	saveDraft, _, err := optionalBoolValue(args, "save_draft")
	if err != nil {
		return err
	}
	includeModelIO, _, err := optionalBoolValue(args, "include_model_io")
	if err != nil {
		return err
	}
	if err := optionalBool(args, "force"); err != nil {
		return err
	}
	if saveDraft {
		if err := requireString(args, "out", 1, 4096); err != nil {
			return fmt.Errorf("out is required when save_draft is true")
		}
	} else if _, hasOut := args["out"]; hasOut {
		return fmt.Errorf("out is valid only when save_draft is true")
	}
	if includeModelIO && !saveDraft {
		return fmt.Errorf("include_model_io requires save_draft")
	}
	return rejectUnknown(args, "memory_id", "platform", "template", "title", "save_draft", "out", "include_model_io", "force")
}

func validateExportWebhookModelIOArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	hasMemory := hasNonEmptyString(args, "memory_id")
	hasRequest := hasNonEmptyString(args, "request_id")
	if hasMemory == hasRequest {
		return fmt.Errorf("exactly one of memory_id or request_id is required")
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
	if err := requireString(args, "out", 1, 4096); err != nil {
		return err
	}
	if err := optionalBool(args, "force"); err != nil {
		return err
	}
	return rejectUnknown(args, "memory_id", "request_id", "platform", "out", "force")
}

func validateSendWebhookArgs(raw json.RawMessage) error {
	args, err := decodeArgs(raw)
	if err != nil {
		return err
	}
	var aliases []string
	if err := requireInto(args, "target_aliases", &aliases); err != nil {
		return err
	}
	if len(aliases) == 0 || len(aliases) > 20 {
		return fmt.Errorf("target_aliases length is out of bounds")
	}
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		normalized, err := webhook.ValidateAlias(alias)
		if err != nil {
			return err
		}
		if _, ok := seen[normalized]; ok {
			return fmt.Errorf("duplicate webhook target alias %q", normalized)
		}
		seen[normalized] = struct{}{}
	}
	if err := optionalString(args, "title", 1, 256); err != nil {
		return err
	}
	if err := optionalString(args, "markdown", 1, maxDirectMarkdownBytes); err != nil {
		return err
	}
	if err := optionalString(args, "draft_dir", 1, 4096); err != nil {
		return err
	}
	if err := optionalString(args, "memory_id", 1, 160); err != nil {
		return err
	}
	if err := optionalPlatformValue(args, "platform"); err != nil {
		return err
	}
	if err := optionalString(args, "template", 1, 512); err != nil {
		return err
	}
	if err := optionalBool(args, "test_message"); err != nil {
		return err
	}
	if err := optionalInt(args, "timeout_seconds", int(minWebhookTimeout/time.Second), int(maxWebhookTimeout/time.Second)); err != nil {
		return err
	}
	saveDraft, _, err := optionalBoolValue(args, "save_draft")
	if err != nil {
		return err
	}
	includeModelIO, _, err := optionalBoolValue(args, "include_model_io")
	if err != nil {
		return err
	}
	if err := optionalBool(args, "force"); err != nil {
		return err
	}
	sourceCount := 0
	if hasNonEmptyString(args, "markdown") {
		sourceCount++
	}
	if hasNonEmptyString(args, "draft_dir") {
		sourceCount++
	}
	if hasNonEmptyString(args, "memory_id") {
		sourceCount++
	}
	if test, ok, err := optionalBoolValue(args, "test_message"); err != nil {
		return err
	} else if ok && test {
		sourceCount++
	}
	if sourceCount != 1 {
		return fmt.Errorf("exactly one webhook message source is required")
	}
	if _, hasPlatform := args["platform"]; hasPlatform && !hasNonEmptyString(args, "memory_id") {
		return fmt.Errorf("platform is valid only with memory_id")
	}
	if saveDraft && !hasNonEmptyString(args, "memory_id") {
		return fmt.Errorf("save_draft is valid only with memory_id")
	}
	if saveDraft {
		if err := requireString(args, "out", 1, 4096); err != nil {
			return fmt.Errorf("out is required when save_draft is true")
		}
	} else if _, hasOut := args["out"]; hasOut {
		return fmt.Errorf("out is valid only when save_draft is true")
	}
	if includeModelIO && !saveDraft {
		return fmt.Errorf("include_model_io requires save_draft")
	}
	return rejectUnknown(args, "target_aliases", "title", "markdown", "draft_dir", "memory_id", "platform", "template", "test_message", "timeout_seconds", "save_draft", "out", "include_model_io", "force")
}

func optionalTargetType(args map[string]json.RawMessage, key string) error {
	raw, ok := args[key]
	if !ok {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s must be a string", key)
	}
	_, err := webhook.ValidateType(value)
	return err
}

func optionalPlatformValue(args map[string]json.RawMessage, key string) error {
	raw, ok := args[key]
	if !ok {
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("%s must be a string", key)
	}
	if value != "mobile" && value != "desktop" {
		return fmt.Errorf("%s must be mobile or desktop", key)
	}
	return nil
}

func optionalBool(args map[string]json.RawMessage, key string) error {
	_, _, err := optionalBoolValue(args, key)
	return err
}

func optionalBoolValue(args map[string]json.RawMessage, key string) (bool, bool, error) {
	raw, ok := args[key]
	if !ok {
		return false, false, nil
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, true, fmt.Errorf("%s must be a boolean", key)
	}
	return value, true, nil
}

func hasNonEmptyString(args map[string]json.RawMessage, key string) bool {
	raw, ok := args[key]
	if !ok {
		return false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return false
	}
	return strings.TrimSpace(value) != ""
}

func absPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("output path is required")
	}
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		abs, err := filepath.Abs(cleaned)
		if err != nil {
			return "", fmt.Errorf("resolve output path: %w", err)
		}
		cleaned = abs
	}
	return cleaned, nil
}

func draftFiles(dir string, includeModelIO bool) map[string]string {
	files := map[string]string{
		"source":   filepath.Join(dir, "source.json"),
		"message":  filepath.Join(dir, "message.md"),
		"metadata": filepath.Join(dir, "metadata.json"),
	}
	if includeModelIO {
		files["model_io"] = filepath.Join(dir, "model-io.json")
	}
	return files
}

func formatToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func isSecretMissing(err error) bool {
	return errors.Is(err, webhook.ErrWebhookSecretMissing) || keychain.IsNotFound(err) || errors.Is(err, os.ErrNotExist)
}

func isUnauthorizedAPIError(err error) bool {
	var apiErr *api.Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == 401
}

func friendlyAgentMCPError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *api.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.StatusCode == 400 && apiErr.Code == "invalid_request" {
		return rpcErr(codeToolError, "invalid_request", "request is invalid; if this memory exists on both mobile and desktop, provide platform mobile or desktop")
	}
	if apiErr.StatusCode == 401 {
		return authRequiredError()
	}
	if apiErr.StatusCode == 404 {
		return rpcErr(codeToolError, "not_found", "record not found or no exportable model IO is available")
	}
	return err
}

func isUserFacingToolError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	for _, marker := range []string{
		"webhook",
		"Markdown",
		"draft",
		"template",
		"output",
		"target",
		"platform",
		"memory",
		"model IO",
		"exactly one",
		"missing",
		"duplicate",
		"disabled",
	} {
		if strings.Contains(message, marker) {
			return true
		}
	}
	return false
}
