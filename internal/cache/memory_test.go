package cache

import (
	"context"
	"testing"
)

type memoryStoreFunc func(ctx context.Context, platform string) ([]Memory, error)

func (f memoryStoreFunc) ListMemories(ctx context.Context, platform string) ([]Memory, error) {
	return f(ctx, platform)
}

func TestSearchMemoriesMatchesAuthorizedMetadata(t *testing.T) {
	store := memoryStoreFunc(func(ctx context.Context, platform string) ([]Memory, error) {
		return []Memory{
			{
				ID:                 "mem_event_1",
				Platform:           "mobile",
				ObjectType:         "event",
				ClientObjectID:     "local_event_1",
				RevisionID:         "rev_1",
				SchemaID:           "patchnote.event.v1",
				SourceAvailability: "metadata_only",
			},
			{
				ID:                 "mem_daily_1",
				Platform:           "desktop",
				ObjectType:         "daily",
				ClientObjectID:     "local_daily_1",
				RevisionID:         "rev_2",
				SchemaID:           "patchnote.daily.v1",
				SourceAvailability: "metadata_only",
			},
		}, nil
	})

	result, err := SearchMemories(context.Background(), store, MemorySearchParams{
		Platform: "mobile",
		Query:    "event",
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("search memories: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "mem_event_1" {
		t.Fatalf("unexpected search result: %+v", result)
	}
}

func TestSearchMemoriesBoundsLimit(t *testing.T) {
	store := memoryStoreFunc(func(ctx context.Context, platform string) ([]Memory, error) {
		return []Memory{
			{ID: "mem_1", Platform: "mobile", SchemaID: "patchnote.event.v1"},
			{ID: "mem_2", Platform: "mobile", SchemaID: "patchnote.event.v1"},
		}, nil
	})

	result, err := SearchMemories(context.Background(), store, MemorySearchParams{
		Platform: "mobile",
		Query:    "event",
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("search memories: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected bounded result, got %+v", result)
	}
}

func TestMemoryIndexUpsertAndSearch(t *testing.T) {
	index := NewMemoryIndex()
	if err := index.UpsertMemories(context.Background(), []Memory{
		{ID: "mem_1", Platform: "mobile", ObjectType: "event", SchemaID: "patchnote.event.v1"},
	}); err != nil {
		t.Fatalf("upsert memories: %v", err)
	}

	result, err := SearchMemories(context.Background(), index, MemorySearchParams{
		Platform: "mobile",
		Query:    "event",
	})
	if err != nil {
		t.Fatalf("search memories: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "mem_1" {
		t.Fatalf("unexpected search result: %+v", result)
	}
}

func TestSearchMemoriesRequiresQueryAndPlatform(t *testing.T) {
	store := memoryStoreFunc(func(ctx context.Context, platform string) ([]Memory, error) {
		return nil, nil
	})

	if _, err := SearchMemories(context.Background(), store, MemorySearchParams{Platform: "mobile"}); err == nil {
		t.Fatal("expected query validation error")
	}
	if _, err := SearchMemories(context.Background(), store, MemorySearchParams{Platform: "agent", Query: "event"}); err == nil {
		t.Fatal("expected platform validation error")
	}
}
