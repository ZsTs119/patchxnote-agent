package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/renderdoc"
	"github.com/ZsTs119/patchxnote-agent/internal/webhook"
	"github.com/spf13/cobra"
)

type webhookTargetOutput struct {
	Alias        string `json:"alias"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	MaskedURL    string `json:"masked_url"`
	Template     string `json:"template,omitempty"`
	SecretStatus string `json:"secret_status"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

func newWebhookCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Configure and send local webhook messages",
	}
	cmd.AddCommand(
		newWebhookSetCommand(state),
		newWebhookListCommand(state),
		newWebhookShowCommand(state),
		newWebhookEnableCommand(state, true),
		newWebhookEnableCommand(state, false),
		newWebhookRemoveCommand(state),
		newWebhookTestCommand(state),
		newWebhookDraftCommand(state),
		newWebhookSendCommand(state),
		newWebhookExportModelIOCommand(state),
	)
	return cmd
}

func newWebhookSetCommand(state *rootState) *cobra.Command {
	var targetType string
	var rawURL string
	var urlStdin bool
	var signingSecret string
	var secretStdin bool
	var clearSecret bool
	var templateName string

	cmd := &cobra.Command{
		Use:   "set <alias>",
		Short: "Create or update a webhook target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if rawURL != "" && urlStdin {
				return fmt.Errorf("--url and --url-stdin are mutually exclusive")
			}
			if signingSecret != "" && secretStdin {
				return fmt.Errorf("--secret and --secret-stdin are mutually exclusive")
			}
			if clearSecret && (signingSecret != "" || secretStdin) {
				return fmt.Errorf("--clear-secret cannot be combined with --secret or --secret-stdin")
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			registry := webhook.NewRegistry(runtime.Config)
			secrets := webhook.NewSecretStore(runtime.Secrets, runtime.Config.Profile)

			alias, err := webhook.ValidateAlias(args[0])
			if err != nil {
				return err
			}
			existing, existingErr := registry.Get(alias)
			exists := existingErr == nil
			if existingErr != nil && !errors.Is(existingErr, webhook.ErrTargetNotFound) {
				return existingErr
			}

			stdinValues, err := readWebhookSetStdin(cmd.InOrStdin(), urlStdin, secretStdin)
			if err != nil {
				return err
			}
			if urlStdin {
				rawURL = stdinValues.url
			}
			if secretStdin {
				signingSecret = stdinValues.secret
			}

			var parsedType webhook.TargetType
			if targetType != "" {
				parsedType, err = webhook.ValidateType(targetType)
				if err != nil {
					return err
				}
			} else if exists {
				parsedType = existing.Type
			} else {
				return fmt.Errorf("--type is required when creating a webhook target")
			}

			maskedURL := existing.MaskedURL
			if rawURL != "" {
				validatedURL, err := webhook.ValidateWebhookURL(rawURL)
				if err != nil {
					return err
				}
				if err := secrets.PutURL(cmd.Context(), alias, validatedURL); err != nil {
					return err
				}
				maskedURL = webhook.MaskWebhookURL(validatedURL)
			} else if !exists {
				return fmt.Errorf("--url or --url-stdin is required when creating a webhook target")
			}

			if signingSecret != "" {
				if err := secrets.PutSigningSecret(cmd.Context(), alias, signingSecret); err != nil {
					return err
				}
			}
			if clearSecret {
				if err := secrets.DeleteSigningSecret(cmd.Context(), alias); err != nil {
					return err
				}
			}

			if templateName == "" && exists {
				templateName = existing.Template
			}
			target, err := registry.Set(alias, parsedType, maskedURL, templateName)
			if err != nil {
				return err
			}
			return writeWebhookTargetResult(cmd, state, target, "ok")
		},
	}
	cmd.Flags().StringVar(&targetType, "type", "", "Webhook target type: feishu, dingtalk, or generic")
	cmd.Flags().StringVar(&rawURL, "url", "", "Webhook URL")
	cmd.Flags().BoolVar(&urlStdin, "url-stdin", false, "Read webhook URL from stdin")
	cmd.Flags().StringVar(&signingSecret, "secret", "", "Optional webhook signing secret")
	cmd.Flags().BoolVar(&secretStdin, "secret-stdin", false, "Read optional signing secret from stdin")
	cmd.Flags().BoolVar(&clearSecret, "clear-secret", false, "Clear the optional signing secret")
	cmd.Flags().StringVar(&templateName, "template", "", "Default template name for this target")
	return cmd
}

func newWebhookListCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured webhook targets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			registry := webhook.NewRegistry(runtime.Config)
			secrets := webhook.NewSecretStore(runtime.Secrets, runtime.Config.Profile)
			output := webhookTargetOutputs(cmd.Context(), registry.List(), secrets)
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				for _, item := range output {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\tenabled=%t\turl=%s\tsecret=%s\n", item.Alias, item.Type, item.Enabled, item.MaskedURL, item.SecretStatus); err != nil {
						return err
					}
				}
				return nil
			case "json":
				return writeJSON(cmd.OutOrStdout(), output)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
}

func newWebhookShowCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "show <alias>",
		Short: "Show one webhook target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			registry := webhook.NewRegistry(runtime.Config)
			target, err := registry.Get(args[0])
			if err != nil {
				return err
			}
			secrets := webhook.NewSecretStore(runtime.Secrets, runtime.Config.Profile)
			return writeWebhookTargetResult(cmd, state, target, webhookSecretStatus(cmd.Context(), secrets, target.Alias))
		},
	}
}

func newWebhookEnableCommand(state *rootState, enabled bool) *cobra.Command {
	use := "enable <alias>"
	short := "Enable a webhook target"
	if !enabled {
		use = "disable <alias>"
		short = "Disable a webhook target"
	}
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			target, err := webhook.NewRegistry(runtime.Config).SetEnabled(args[0], enabled)
			if err != nil {
				return err
			}
			status := "enabled"
			if !enabled {
				status = "disabled"
			}
			return writeWebhookTargetResult(cmd, state, target, status)
		},
	}
}

func newWebhookRemoveCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove a webhook target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			registry := webhook.NewRegistry(runtime.Config)
			target, err := registry.Remove(args[0])
			if err != nil {
				return err
			}
			secrets := webhook.NewSecretStore(runtime.Secrets, runtime.Config.Profile)
			if err := secrets.DeleteTarget(cmd.Context(), target.Alias); err != nil {
				return err
			}
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "removed webhook target %s\n", target.Alias)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]any{"removed": true, "alias": target.Alias})
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
}

func newWebhookTestCommand(state *rootState) *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "test <alias>",
		Short: "Send a small test message to a webhook target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			resolved, err := resolveTargets(cmd.Context(), runtime, []string{args[0]})
			if err != nil {
				return err
			}
			message := webhook.Message{
				Title:    "PatchXNote Webhook Test",
				Markdown: "# PatchXNote Webhook Test\n\n这是一条 PatchXNote Agent 本地测试消息。",
				Metadata: map[string]string{"source": "test"},
			}
			results, sendErr := webhook.NewSender(nil).Send(cmd.Context(), resolved, message, webhook.SendOptions{Timeout: timeout})
			if err := writeWebhookSendResults(cmd, state, results); err != nil {
				return err
			}
			return sendErr
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", webhook.DefaultSendTimeout, "Webhook request timeout")
	return cmd
}

type webhookSetStdin struct {
	url    string
	secret string
}

func readWebhookSetStdin(reader io.Reader, needURL bool, needSecret bool) (webhookSetStdin, error) {
	var values webhookSetStdin
	if !needURL && !needSecret {
		return values, nil
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return values, fmt.Errorf("read stdin: %w", err)
	}
	text := strings.TrimSpace(string(body))
	if needURL && needSecret {
		lines := strings.SplitN(text, "\n", 2)
		if len(lines) < 2 {
			return values, fmt.Errorf("--url-stdin and --secret-stdin require two stdin lines")
		}
		values.url = strings.TrimSpace(lines[0])
		values.secret = strings.TrimSpace(lines[1])
		return values, nil
	}
	if needURL {
		values.url = text
	}
	if needSecret {
		values.secret = text
	}
	return values, nil
}

func webhookTargetOutputs(ctx context.Context, targets []webhook.Target, secrets *webhook.SecretStore) []webhookTargetOutput {
	output := make([]webhookTargetOutput, 0, len(targets))
	for _, target := range targets {
		output = append(output, webhookTargetOutputFromTarget(target, webhookSecretStatus(ctx, secrets, target.Alias)))
	}
	return output
}

func webhookTargetOutputFromTarget(target webhook.Target, secretStatus string) webhookTargetOutput {
	return webhookTargetOutput{
		Alias:        target.Alias,
		Type:         string(target.Type),
		Enabled:      target.Enabled,
		MaskedURL:    target.MaskedURL,
		Template:     target.Template,
		SecretStatus: secretStatus,
		CreatedAt:    formatOutputTime(target.CreatedAt),
		UpdatedAt:    formatOutputTime(target.UpdatedAt),
	}
}

func writeWebhookTargetResult(cmd *cobra.Command, state *rootState, target webhook.Target, secretStatus string) error {
	output := webhookTargetOutputFromTarget(target, secretStatus)
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "alias %s\ntype %s\nenabled %t\nurl %s\nsecret %s\n", output.Alias, output.Type, output.Enabled, output.MaskedURL, output.SecretStatus)
		return err
	case "json":
		return writeJSON(cmd.OutOrStdout(), output)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func webhookSecretStatus(ctx context.Context, secrets *webhook.SecretStore, alias string) string {
	if _, err := secrets.URL(ctx, alias); err != nil {
		if errors.Is(err, webhook.ErrWebhookSecretMissing) {
			return "missing"
		}
		return "unavailable"
	}
	return "ok"
}

func formatOutputTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func newWebhookDraftCommand(state *rootState) *cobra.Command {
	var memoryID string
	var platform string
	var templateName string
	var outDir string
	var includeModelIO bool
	var force bool
	cmd := &cobra.Command{
		Use:   "draft",
		Short: "Render a PatchXNote memory into local webhook draft files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(memoryID) == "" {
				return fmt.Errorf("--memory-id is required")
			}
			if strings.TrimSpace(outDir) == "" {
				return fmt.Errorf("--out is required")
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			delivery, err := fetchDeliveryDocument(cmd.Context(), runtime, platform, memoryID)
			if err != nil {
				return err
			}
			doc := renderdoc.FromDeliveryDocument(delivery)
			message, err := renderdoc.RenderTemplate(templateName, doc)
			if err != nil {
				return err
			}
			var modelIO *api.AgentModelIOExport
			if includeModelIO {
				got, err := fetchMemoryModelIO(cmd.Context(), runtime, platform, memoryID)
				if err != nil {
					return err
				}
				modelIO = &got
			}
			if err := writeDraftDirectory(outDir, delivery, message, templateNameOrDefault(templateName), modelIO, force); err != nil {
				return err
			}
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "draft written\nout %s\n", outDir)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]any{"draft_written": true, "out": outDir})
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
	cmd.Flags().StringVar(&memoryID, "memory-id", "", "PatchXNote memory ID")
	cmd.Flags().StringVar(&platform, "platform", "", "Optional memory platform: mobile or desktop")
	cmd.Flags().StringVar(&templateName, "template", "default", "Template name or local template file path")
	cmd.Flags().StringVar(&outDir, "out", "", "Draft output directory")
	cmd.Flags().BoolVar(&includeModelIO, "include-model-io", false, "Also export explicit model IO JSON into the draft directory")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite known draft files")
	return cmd
}

func newWebhookSendCommand(state *rootState) *cobra.Command {
	var targets []string
	var filePath string
	var draftDir string
	var memoryID string
	var platform string
	var templateName string
	var title string
	var saveDraft bool
	var outDir string
	var includeModelIO bool
	var force bool
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send Markdown to one or more configured webhook targets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(targets) == 0 {
				return fmt.Errorf("at least one --target is required")
			}
			sourceCount := countNonEmpty(filePath, draftDir, memoryID)
			if sourceCount != 1 {
				return fmt.Errorf("exactly one of --file, --draft, or --memory-id is required")
			}
			if platform != "" && memoryID == "" {
				return fmt.Errorf("--platform is valid only with --memory-id")
			}
			if saveDraft && memoryID == "" {
				return fmt.Errorf("--save-draft is valid only with --memory-id")
			}
			if includeModelIO && !saveDraft {
				return fmt.Errorf("--include-model-io requires --save-draft")
			}
			if saveDraft && outDir == "" {
				return fmt.Errorf("--out is required with --save-draft")
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			message, err := buildWebhookMessage(cmd.Context(), runtime, filePath, draftDir, memoryID, platform, templateName, title, saveDraft, outDir, includeModelIO, force)
			if err != nil {
				return err
			}
			resolved, err := resolveTargets(cmd.Context(), runtime, targets)
			if err != nil {
				return err
			}
			results, sendErr := webhook.NewSender(nil).Send(cmd.Context(), resolved, message, webhook.SendOptions{Timeout: timeout})
			if err := writeWebhookSendResults(cmd, state, results); err != nil {
				return err
			}
			return sendErr
		},
	}
	cmd.Flags().StringArrayVar(&targets, "target", nil, "Webhook target alias; may be repeated")
	cmd.Flags().StringVar(&filePath, "file", "", "Markdown file to send")
	cmd.Flags().StringVar(&draftDir, "draft", "", "Draft directory containing message.md")
	cmd.Flags().StringVar(&memoryID, "memory-id", "", "PatchXNote memory ID")
	cmd.Flags().StringVar(&platform, "platform", "", "Optional memory platform: mobile or desktop")
	cmd.Flags().StringVar(&templateName, "template", "default", "Template name or local template file path")
	cmd.Flags().StringVar(&title, "title", "", "Override title for --file")
	cmd.Flags().BoolVar(&saveDraft, "save-draft", false, "Write memory-backed draft files before sending")
	cmd.Flags().StringVar(&outDir, "out", "", "Draft output directory for --save-draft")
	cmd.Flags().BoolVar(&includeModelIO, "include-model-io", false, "Also save model-io.json with --save-draft")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite known draft files")
	cmd.Flags().DurationVar(&timeout, "timeout", webhook.DefaultSendTimeout, "Webhook request timeout")
	return cmd
}

func newWebhookExportModelIOCommand(state *rootState) *cobra.Command {
	var memoryID string
	var requestID string
	var platform string
	var outFile string
	var force bool
	cmd := &cobra.Command{
		Use:   "export-model-io",
		Short: "Export explicit Agent model IO JSON to a local file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if countNonEmpty(memoryID, requestID) != 1 {
				return fmt.Errorf("exactly one of --memory-id or --request-id is required")
			}
			if outFile == "" {
				return fmt.Errorf("--out is required")
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			var export api.AgentModelIOExport
			if memoryID != "" {
				export, err = fetchMemoryModelIO(cmd.Context(), runtime, platform, memoryID)
			} else {
				export, err = fetchModelRunIOTrace(cmd.Context(), runtime, platform, requestID)
			}
			if err != nil {
				return err
			}
			body, err := json.MarshalIndent(export, "", "  ")
			if err != nil {
				return fmt.Errorf("encode model IO export: %w", err)
			}
			body = append(body, '\n')
			if err := atomicWriteFile(outFile, body, 0o600, force); err != nil {
				return err
			}
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "model io exported\nout %s\n", outFile)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]any{"exported": true, "out": outFile})
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
	cmd.Flags().StringVar(&memoryID, "memory-id", "", "PatchXNote memory ID")
	cmd.Flags().StringVar(&requestID, "request-id", "", "PatchXNote model request/run ID")
	cmd.Flags().StringVar(&platform, "platform", "", "Optional platform: mobile or desktop")
	cmd.Flags().StringVar(&outFile, "out", "", "Output JSON file")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing output file")
	return cmd
}

func buildWebhookMessage(ctx context.Context, runtime runtimeState, filePath string, draftDir string, memoryID string, platform string, templateName string, title string, saveDraft bool, outDir string, includeModelIO bool, force bool) (webhook.Message, error) {
	switch {
	case filePath != "":
		return messageFromFile(filePath, title)
	case draftDir != "":
		return messageFromDraft(draftDir)
	case memoryID != "":
		delivery, err := fetchDeliveryDocument(ctx, runtime, platform, memoryID)
		if err != nil {
			return webhook.Message{}, err
		}
		doc := renderdoc.FromDeliveryDocument(delivery)
		rendered, err := renderdoc.RenderTemplate(templateName, doc)
		if err != nil {
			return webhook.Message{}, err
		}
		var modelIO *api.AgentModelIOExport
		if includeModelIO {
			got, err := fetchMemoryModelIO(ctx, runtime, platform, memoryID)
			if err != nil {
				return webhook.Message{}, err
			}
			modelIO = &got
		}
		if saveDraft {
			if err := writeDraftDirectory(outDir, delivery, rendered, templateNameOrDefault(templateName), modelIO, force); err != nil {
				return webhook.Message{}, err
			}
		}
		return messageFromDelivery(delivery, rendered, templateNameOrDefault(templateName)), nil
	default:
		return webhook.Message{}, fmt.Errorf("missing webhook message source")
	}
}

func messageFromFile(path string, titleOverride string) (webhook.Message, error) {
	body, err := readMarkdownFile(path)
	if err != nil {
		return webhook.Message{}, err
	}
	return webhook.Message{
		Title:    renderdoc.InferTitle(string(body), path, titleOverride),
		Markdown: string(body),
		Metadata: map[string]string{"source": "file"},
	}, nil
}

func messageFromDraft(dir string) (webhook.Message, error) {
	messagePath := filepath.Join(dir, "message.md")
	body, err := readMarkdownFile(messagePath)
	if err != nil {
		return webhook.Message{}, err
	}
	metadata := readDraftMetadata(filepath.Join(dir, "metadata.json"))
	message := webhook.Message{
		Title:    renderdoc.InferTitle(string(body), dir, ""),
		Markdown: string(body),
		Metadata: metadata,
	}
	if metadata["memory_id"] != "" || metadata["platform"] != "" {
		message.Memory = &webhook.MessageMemory{ID: metadata["memory_id"], Platform: metadata["platform"]}
	}
	return message, nil
}

func messageFromDelivery(delivery api.AgentDeliveryDocument, markdown string, templateName string) webhook.Message {
	message := webhook.Message{
		Title:    delivery.Title,
		Markdown: markdown,
		Metadata: map[string]string{
			"source":   "patchxnote",
			"template": templateName,
		},
	}
	if delivery.Memory != nil {
		message.Memory = &webhook.MessageMemory{ID: delivery.Memory.ID, Platform: delivery.Memory.Platform}
	}
	return message
}

func readMarkdownFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read markdown file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("markdown path must be a file")
	}
	if info.Size() > renderdoc.MaxRenderedMarkdownBytes {
		return nil, fmt.Errorf("Markdown file exceeds local safety cap")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read markdown file: %w", err)
	}
	if len(body) > renderdoc.MaxRenderedMarkdownBytes {
		return nil, fmt.Errorf("Markdown file exceeds local safety cap")
	}
	return body, nil
}

func readDraftMetadata(path string) map[string]string {
	body, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{"source": "draft"}
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return map[string]string{"source": "draft"}
	}
	metadata := make(map[string]string, len(raw)+1)
	metadata["source"] = "draft"
	for key, value := range raw {
		switch typed := value.(type) {
		case string:
			metadata[key] = typed
		case bool:
			metadata[key] = fmt.Sprintf("%t", typed)
		case float64:
			metadata[key] = fmt.Sprintf("%.0f", typed)
		}
	}
	return metadata
}

func resolveTargets(ctx context.Context, runtime runtimeState, aliases []string) ([]webhook.ResolvedTarget, error) {
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

	registry := webhook.NewRegistry(runtime.Config)
	secrets := webhook.NewSecretStore(runtime.Secrets, runtime.Config.Profile)
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

func fetchDeliveryDocument(ctx context.Context, runtime runtimeState, platform string, memoryID string) (api.AgentDeliveryDocument, error) {
	var delivery api.AgentDeliveryDocument
	err := withAgentAccessToken(ctx, runtime, func(accessToken string) error {
		var err error
		delivery, err = runtime.API.GetMemoryDeliveryDocument(ctx, accessToken, platform, memoryID)
		return err
	})
	return delivery, err
}

func fetchMemoryModelIO(ctx context.Context, runtime runtimeState, platform string, memoryID string) (api.AgentModelIOExport, error) {
	var export api.AgentModelIOExport
	err := withAgentAccessToken(ctx, runtime, func(accessToken string) error {
		var err error
		export, err = runtime.API.GetMemoryModelIO(ctx, accessToken, platform, memoryID)
		return err
	})
	return export, err
}

func fetchModelRunIOTrace(ctx context.Context, runtime runtimeState, platform string, requestID string) (api.AgentModelIOExport, error) {
	var export api.AgentModelIOExport
	err := withAgentAccessToken(ctx, runtime, func(accessToken string) error {
		var err error
		export, err = runtime.API.GetModelRunIOTrace(ctx, accessToken, platform, requestID)
		return err
	})
	return export, err
}

func withAgentAccessToken(ctx context.Context, runtime runtimeState, call func(accessToken string) error) error {
	if runtime.API == nil {
		return fmt.Errorf("PatchXNote API client is not configured")
	}
	credential, ok, err := runtime.Credentials.Credential(ctx)
	if err != nil {
		return err
	}
	if !ok || credential.AccessToken == "" {
		return fmt.Errorf("agent login required; run patchxnote login")
	}
	err = call(credential.AccessToken)
	if isUnauthorizedAPIError(err) && credential.RefreshToken != "" {
		refreshed, refreshedOK, refreshErr := runtime.Credentials.RefreshNow(ctx)
		if refreshErr != nil {
			return friendlyAgentAPIError(refreshErr)
		}
		if refreshedOK && refreshed.AccessToken != "" {
			err = call(refreshed.AccessToken)
		}
	}
	return friendlyAgentAPIError(err)
}

func isUnauthorizedAPIError(err error) bool {
	var apiErr *api.Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == 401
}

func friendlyAgentAPIError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *api.Error
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.StatusCode == 400 && apiErr.Code == "invalid_request" {
		return fmt.Errorf("request is invalid; if this memory exists on both mobile and desktop, rerun with --platform mobile or --platform desktop")
	}
	if apiErr.StatusCode == 401 {
		return fmt.Errorf("agent login required; run patchxnote login")
	}
	if apiErr.StatusCode == 404 {
		return fmt.Errorf("record not found or no exportable model IO is available")
	}
	return err
}

func writeDraftDirectory(dir string, delivery api.AgentDeliveryDocument, message string, templateName string, modelIO *api.AgentModelIOExport, force bool) error {
	if dir == "" {
		return fmt.Errorf("draft output directory is required")
	}
	if info, err := os.Lstat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("draft output path must be a directory")
		}
		if !force {
			return fmt.Errorf("draft output directory already exists; pass --force to overwrite known draft files")
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create draft output directory: %w", err)
		}
	} else {
		return fmt.Errorf("inspect draft output directory: %w", err)
	}

	sourceJSON, err := json.MarshalIndent(delivery, "", "  ")
	if err != nil {
		return fmt.Errorf("encode delivery source: %w", err)
	}
	sourceJSON = append(sourceJSON, '\n')
	if err := atomicWriteFile(filepath.Join(dir, "source.json"), sourceJSON, 0o600, force); err != nil {
		return err
	}
	if err := atomicWriteFile(filepath.Join(dir, "message.md"), []byte(message), 0o600, force); err != nil {
		return err
	}
	metadataJSON, err := json.MarshalIndent(draftMetadata(delivery, templateName, modelIO != nil), "", "  ")
	if err != nil {
		return fmt.Errorf("encode draft metadata: %w", err)
	}
	metadataJSON = append(metadataJSON, '\n')
	if err := atomicWriteFile(filepath.Join(dir, "metadata.json"), metadataJSON, 0o600, force); err != nil {
		return err
	}
	if modelIO != nil {
		body, err := json.MarshalIndent(modelIO, "", "  ")
		if err != nil {
			return fmt.Errorf("encode model IO: %w", err)
		}
		body = append(body, '\n')
		if err := atomicWriteFile(filepath.Join(dir, "model-io.json"), body, 0o600, force); err != nil {
			return err
		}
	}
	return nil
}

func draftMetadata(delivery api.AgentDeliveryDocument, templateName string, modelIOIncluded bool) map[string]any {
	metadata := map[string]any{
		"source":              "patchxnote",
		"version":             webhook.PatchXNoteWebhookPayloadVersion,
		"template":            templateName,
		"delivery_request_id": delivery.Trace.RequestID,
		"model_io_included":   modelIOIncluded,
		"generated_at":        time.Now().UTC().Format(time.RFC3339),
	}
	if delivery.Memory != nil {
		metadata["platform"] = delivery.Memory.Platform
		metadata["memory_id"] = delivery.Memory.ID
	} else if delivery.Trace.Platform != "" {
		metadata["platform"] = delivery.Trace.Platform
	}
	return metadata
}

func atomicWriteFile(path string, body []byte, perm os.FileMode, overwrite bool) error {
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
			return fmt.Errorf("refusing to overwrite non-regular output path %s", path)
		}
		if !overwrite {
			return fmt.Errorf("output file already exists: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temporary output file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("set output file permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary output file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}
	return nil
}

func writeWebhookSendResults(cmd *cobra.Command, state *rootState, results []webhook.SendResult) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		for _, result := range results {
			if result.Success {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "sent %s status=%d\n", result.Alias, result.StatusCode); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "failed %s status=%d code=%s message=%s error=%s\n", result.Alias, result.StatusCode, result.ProviderCode, result.ProviderMessage, result.Error); err != nil {
				return err
			}
		}
		return nil
	case "json":
		return writeJSON(cmd.OutOrStdout(), results)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func countNonEmpty(values ...string) int {
	var count int
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	return count
}

func templateNameOrDefault(value string) string {
	if strings.TrimSpace(value) == "" {
		return "default"
	}
	return value
}

var _ keychain.SecretStore = keychain.UnavailableStore{}
