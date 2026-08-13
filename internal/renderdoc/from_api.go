package renderdoc

import "github.com/ZsTs119/patchxnote-agent/internal/api"

func FromDeliveryDocument(source api.AgentDeliveryDocument) Document {
	doc := Document{
		Title:       source.Title,
		Summary:     source.Summary,
		Markdown:    source.Markdown,
		Source:      source.Source,
		Version:     source.Version,
		GeneratedAt: source.GeneratedAt,
		Sections:    make([]Section, 0, len(source.Sections)),
		KeyItems:    make([]KeyItem, 0, len(source.KeyItems)),
		Metadata:    make(map[string]string),
		Trace: TraceRef{
			RequestID: source.Trace.RequestID,
			Platform:  source.Trace.Platform,
			TaskType:  source.Trace.TaskType,
			State:     source.Trace.State,
		},
	}
	for _, section := range source.Sections {
		doc.Sections = append(doc.Sections, Section{
			Title:    section.Title,
			Markdown: section.Markdown,
		})
	}
	for _, item := range source.KeyItems {
		doc.KeyItems = append(doc.KeyItems, KeyItem{
			Title:    item.Title,
			Status:   item.Status,
			Owner:    item.Owner,
			DueAt:    item.DueAt,
			Markdown: item.Markdown,
		})
	}
	if source.Memory != nil {
		doc.Memory = MemoryRef{
			ID:         source.Memory.ID,
			Platform:   source.Memory.Platform,
			ObjectType: source.Memory.ObjectType,
			RevisionID: source.Memory.RevisionID,
		}
	}
	if doc.Source != "" {
		doc.Metadata["source"] = doc.Source
	}
	if doc.Version != "" {
		doc.Metadata["version"] = doc.Version
	}
	return doc
}
