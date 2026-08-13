package renderdoc

import "time"

type Document struct {
	Title       string
	Summary     string
	Markdown    string
	Sections    []Section
	KeyItems    []KeyItem
	Memory      MemoryRef
	Trace       TraceRef
	Source      string
	Version     string
	GeneratedAt time.Time
	Metadata    map[string]string
}

type Section struct {
	Title    string
	Markdown string
}

type KeyItem struct {
	Title    string
	Status   string
	Owner    string
	DueAt    string
	Markdown string
}

type MemoryRef struct {
	ID         string `json:"id,omitempty"`
	Platform   string `json:"platform,omitempty"`
	ObjectType string `json:"object_type,omitempty"`
	RevisionID string `json:"revision_id,omitempty"`
}

type TraceRef struct {
	RequestID string `json:"request_id,omitempty"`
	Platform  string `json:"platform,omitempty"`
	TaskType  string `json:"task_type,omitempty"`
	State     string `json:"state,omitempty"`
}
