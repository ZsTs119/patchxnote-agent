package webhook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/localfile"
	"github.com/ZsTs119/patchxnote-agent/internal/renderdoc"
)

func MessageFromFile(path string, titleOverride string) (Message, error) {
	body, err := ReadMarkdownFile(path)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Title:    renderdoc.InferTitle(string(body), path, titleOverride),
		Markdown: string(body),
		Metadata: map[string]string{"source": "file"},
	}, nil
}

func MessageFromMarkdown(markdown string, titleOverride string) (Message, error) {
	if len(markdown) > renderdoc.MaxRenderedMarkdownBytes {
		return Message{}, fmt.Errorf("Markdown input exceeds local safety cap")
	}
	return Message{
		Title:    renderdoc.InferTitle(markdown, "PatchXNote 记录", titleOverride),
		Markdown: markdown,
		Metadata: map[string]string{"source": "direct"},
	}, nil
}

func MessageFromDraft(dir string) (Message, error) {
	messagePath := filepath.Join(dir, "message.md")
	body, err := ReadMarkdownFile(messagePath)
	if err != nil {
		return Message{}, err
	}
	metadata := ReadDraftMetadata(filepath.Join(dir, "metadata.json"))
	message := Message{
		Title:    renderdoc.InferTitle(string(body), dir, ""),
		Markdown: string(body),
		Metadata: metadata,
	}
	if metadata["memory_id"] != "" || metadata["platform"] != "" {
		message.Memory = &MessageMemory{ID: metadata["memory_id"], Platform: metadata["platform"]}
	}
	return message, nil
}

func MessageFromDelivery(delivery api.AgentDeliveryDocument, markdown string, templateName string) Message {
	message := Message{
		Title:    delivery.Title,
		Markdown: markdown,
		Metadata: map[string]string{
			"source":   "patchxnote",
			"template": TemplateNameOrDefault(templateName),
		},
	}
	if delivery.Memory != nil {
		message.Memory = &MessageMemory{ID: delivery.Memory.ID, Platform: delivery.Memory.Platform}
	}
	return message
}

func ReadMarkdownFile(path string) ([]byte, error) {
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

func ReadDraftMetadata(path string) map[string]string {
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

func WriteDraftDirectory(dir string, delivery api.AgentDeliveryDocument, message string, templateName string, modelIO *api.AgentModelIOExport, force bool) error {
	if dir == "" {
		return fmt.Errorf("draft output directory is required")
	}
	if info, err := os.Lstat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("draft output path must be a directory")
		}
		if !force {
			return fmt.Errorf("draft output directory already exists; pass force to overwrite known draft files")
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
	if err := AtomicWriteFile(filepath.Join(dir, "source.json"), sourceJSON, 0o600, force); err != nil {
		return err
	}
	if err := AtomicWriteFile(filepath.Join(dir, "message.md"), []byte(message), 0o600, force); err != nil {
		return err
	}
	metadataJSON, err := json.MarshalIndent(DraftMetadata(delivery, templateName, modelIO != nil), "", "  ")
	if err != nil {
		return fmt.Errorf("encode draft metadata: %w", err)
	}
	metadataJSON = append(metadataJSON, '\n')
	if err := AtomicWriteFile(filepath.Join(dir, "metadata.json"), metadataJSON, 0o600, force); err != nil {
		return err
	}
	if modelIO != nil {
		body, err := json.MarshalIndent(modelIO, "", "  ")
		if err != nil {
			return fmt.Errorf("encode model IO: %w", err)
		}
		body = append(body, '\n')
		if err := AtomicWriteFile(filepath.Join(dir, "model-io.json"), body, 0o600, force); err != nil {
			return err
		}
	}
	return nil
}

func DraftMetadata(delivery api.AgentDeliveryDocument, templateName string, modelIOIncluded bool) map[string]any {
	metadata := map[string]any{
		"source":              "patchxnote",
		"version":             PatchXNoteWebhookPayloadVersion,
		"template":            TemplateNameOrDefault(templateName),
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

func AtomicWriteFile(path string, body []byte, perm os.FileMode, overwrite bool) error {
	return localfile.AtomicWriteFile(path, body, perm, overwrite)
}

func TemplateNameOrDefault(templateName string) string {
	if templateName == "" {
		return "default"
	}
	return templateName
}
