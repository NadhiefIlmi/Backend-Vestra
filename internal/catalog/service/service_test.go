package service

import (
	"context"
	"errors"
	"testing"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/repository"
)

func TestServiceListClothing(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := New(repo)
	ctx := context.Background()

	// 1. Fetch all published items
	res, err := svc.ListClothing(ctx, ListClothingParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) == 0 {
		t.Errorf("expected items in default catalog, got 0")
	}
	if res.Page != 1 || res.Limit != 10 {
		t.Errorf("unexpected pagination params: page %d, limit %d", res.Page, res.Limit)
	}

	// 2. Fetch with category filter
	topRes, err := svc.ListClothing(ctx, ListClothingParams{Category: "top", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error with category filter: %v", err)
	}
	for _, it := range topRes.Items {
		if it.Category != "top" {
			t.Errorf("expected category 'top', got '%s'", it.Category)
		}
	}

	// 3. Test pagination limit capping
	cappedRes, err := svc.ListClothing(ctx, ListClothingParams{Limit: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cappedRes.Limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", cappedRes.Limit)
	}
}

func TestServiceGetClothingByID(t *testing.T) {
	repo := repository.NewMemoryRepository()
	svc := New(repo)
	ctx := context.Background()

	// Existing item
	item, err := svc.GetClothingByID(ctx, "shirt-basic-black")
	if err != nil {
		t.Fatalf("expected item to be found, got err: %v", err)
	}
	if item.ID != "shirt-basic-black" {
		t.Errorf("expected id shirt-basic-black, got %s", item.ID)
	}

	// Non-existing item
	_, err = svc.GetClothingByID(ctx, "non-existent-id")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// Empty ID
	_, err = svc.GetClothingByID(ctx, "")
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("expected ErrInvalidID for empty id, got %v", err)
	}
}

