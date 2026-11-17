package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/model"
)

func TestMemoryRepository_FindPublished(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	t.Run("returns all seeded published items with default pagination", func(t *testing.T) {
		items, total, err := repo.FindPublished(ctx, "", 0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total < 3 {
			t.Errorf("expected at least 3 total items, got %d", total)
		}
		if len(items) != total {
			t.Errorf("expected %d items returned, got %d", total, len(items))
		}
	})

	t.Run("filters by category", func(t *testing.T) {
		items, total, err := repo.FindPublished(ctx, "top", 0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total == 0 {
			t.Fatal("expected at least 1 top category item")
		}
		for _, item := range items {
			if item.Category != "top" {
				t.Errorf("expected category 'top', got '%s'", item.Category)
			}
		}
	})

	t.Run("returns empty slice for non-existent category", func(t *testing.T) {
		items, total, err := repo.FindPublished(ctx, "non_existent_cat", 0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 || len(items) != 0 {
			t.Errorf("expected 0 items for unknown category, got %d (total: %d)", len(items), total)
		}
	})

	t.Run("respects pagination offset and limit", func(t *testing.T) {
		items, total, err := repo.FindPublished(ctx, "", 0, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item with limit=1, got %d", len(items))
		}
		if total < 2 {
			t.Errorf("expected total >= 2, got %d", total)
		}

		offsetItems, _, err := repo.FindPublished(ctx, "", 100, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(offsetItems) != 0 {
			t.Errorf("expected 0 items with offset > total, got %d", len(offsetItems))
		}
	})

	t.Run("returns context error when cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, err := repo.FindPublished(cancelCtx, "", 0, 10)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestMemoryRepository_FindByID(t *testing.T) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	t.Run("finds existing item by ID", func(t *testing.T) {
		item, err := repo.FindByID(ctx, "shirt-basic-black")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item == nil || item.ID != "shirt-basic-black" {
			t.Fatalf("expected item 'shirt-basic-black', got %v", item)
		}
		if item.Name != "Classic Black T-Shirt" {
			t.Errorf("expected name 'Classic Black T-Shirt', got '%s'", item.Name)
		}
	})

	t.Run("returns ErrNotFound for non-existing ID", func(t *testing.T) {
		item, err := repo.FindByID(ctx, "non-existent-id")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
		if item != nil {
			t.Errorf("expected nil item on error, got %v", item)
		}
	})

	t.Run("returns context error when cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.FindByID(cancelCtx, "shirt-basic-black")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("does not return unpublished item", func(t *testing.T) {
		repo.mu.Lock()
		repo.items["draft-item"] = model.ClothingItem{
			ID:        "draft-item",
			Name:      "Draft Item",
			Status:    model.StatusDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.mu.Unlock()

		item, err := repo.FindByID(ctx, "draft-item")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound for draft item, got %v", err)
		}
		if item != nil {
			t.Errorf("expected nil item, got %v", item)
		}
	})
}

