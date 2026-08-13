package renderdoc

import (
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
)

func TestFromDeliveryDocumentMapsSafeProjection(t *testing.T) {
	generatedAt := time.Date(2026, 8, 13, 8, 2, 0, 0, time.UTC)
	source := api.AgentDeliveryDocument{
		Source:      "patchxnote",
		Version:     "1",
		Title:       "合成会议纪要",
		Summary:     "合成摘要",
		Markdown:    "# 合成会议纪要\n\n正文",
		GeneratedAt: generatedAt,
		Sections: []api.AgentDeliverySection{
			{Title: "关键结论", Markdown: "- 结论"},
		},
		KeyItems: []api.AgentDeliveryKeyItem{
			{Title: "完成联调", Status: "open", Markdown: "- 完成联调"},
		},
		Memory: &api.AgentDeliveryMemory{
			ID:         "mem_fixture",
			Platform:   "desktop",
			ObjectType: "daily_digest",
			RevisionID: "rev_fixture",
		},
		Trace: api.AgentDeliveryTrace{
			RequestID: "mrun_fixture",
			Platform:  "desktop",
			TaskType:  "daily_digest",
			State:     "completed",
		},
	}

	doc := FromDeliveryDocument(source)
	if doc.Title != source.Title || doc.GeneratedAt != generatedAt || len(doc.Sections) != 1 || len(doc.KeyItems) != 1 {
		t.Fatalf("unexpected document mapping: %+v", doc)
	}
	if doc.Memory.ID != "mem_fixture" || doc.Trace.RequestID != "mrun_fixture" {
		t.Fatalf("unexpected refs: memory=%+v trace=%+v", doc.Memory, doc.Trace)
	}
	if doc.Metadata["source"] != "patchxnote" || doc.Metadata["version"] != "1" {
		t.Fatalf("expected bounded metadata, got %+v", doc.Metadata)
	}
}

func TestFromDeliveryDocumentHandlesMissingOptionalMemory(t *testing.T) {
	doc := FromDeliveryDocument(api.AgentDeliveryDocument{
		Title: "无 memory",
		Trace: api.AgentDeliveryTrace{RequestID: "mrun_fixture"},
	})
	if doc.Memory.ID != "" || doc.Title != "无 memory" {
		t.Fatalf("unexpected missing memory mapping: %+v", doc)
	}
}
