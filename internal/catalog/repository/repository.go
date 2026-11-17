package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/model"
)

var (
	ErrNotFound = errors.New("clothing item not found")
)

type Repository interface {
	FindPublished(ctx context.Context, category string, offset, limit int) ([]model.ClothingItem, int, error)
	FindByID(ctx context.Context, id string) (*model.ClothingItem, error)
}

// MemoryRepository provides an in-memory implementation for tests and dev seeding
type MemoryRepository struct {
	mu    sync.RWMutex
	items map[string]model.ClothingItem
}

func NewMemoryRepository() *MemoryRepository {
	now := time.Now()
	repo := &MemoryRepository{
		items: make(map[string]model.ClothingItem),
	}

	// Seed with default catalog items matching PRD and frontend
	items := []model.ClothingItem{
		{
			ID:           "shirt-joni-vintage",
			Name:         "Joni Vintage Graphic T-Shirt",
			Description:  "Kaos hitam vintage graphic print \"JONI\" bertekstur kain katun nyata.",
			Category:     "top",
			ThumbnailURL: "/assets/clothing/joni-vintage-black.png",
			AssetURL:     "/assets/clothing/joni-vintage-black.png",
			AssetMetadata: model.AssetMetadata{
				Anchor:          model.AnchorConfig{X: 0.5, Y: 0.28},
				ReferenceWidth:  800,
				ScaleMultiplier: 1.15,
				RotationOffset:  0,
				Offset:          model.OffsetConfig{X: 0, Y: 0},
			},
			Status:    model.StatusPublished,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "shirt-basic-black",
			Name:         "Classic Black T-Shirt",
			Description:  "Kaos hitam katun combed 30s premium dengan fitting regular modern.",
			Category:     "top",
			ThumbnailURL: "/assets/clothing/shirt-black-thumb.webp",
			AssetURL:     "/assets/clothing/shirt-black.png",
			AssetMetadata: model.AssetMetadata{
				Anchor:          model.AnchorConfig{X: 0.5, Y: 0.28},
				ReferenceWidth:  1000,
				ScaleMultiplier: 1.15,
				RotationOffset:  0,
				Offset:          model.OffsetConfig{X: 0, Y: 0},
			},
			Status:    model.StatusPublished,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "shirt-basic-white",
			Name:         "Classic White T-Shirt",
			Description:  "Kaos putih netral dengan kerah bulat lembut, pas untuk daily wear.",
			Category:     "top",
			ThumbnailURL: "/assets/clothing/shirt-white-thumb.webp",
			AssetURL:     "/assets/clothing/shirt-white.png",
			AssetMetadata: model.AssetMetadata{
				Anchor:          model.AnchorConfig{X: 0.5, Y: 0.28},
				ReferenceWidth:  1000,
				ScaleMultiplier: 1.15,
				RotationOffset:  0,
				Offset:          model.OffsetConfig{X: 0, Y: 0},
			},
			Status:    model.StatusPublished,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           "hoodie-urban-grey",
			Name:         "Urban Oversized Hoodie",
			Description:  "Hoodie abu-abu streetwear fleece tebal dengan potongan santai.",
			Category:     "top",
			ThumbnailURL: "/assets/clothing/hoodie-grey-thumb.webp",
			AssetURL:     "/assets/clothing/hoodie-grey.png",
			AssetMetadata: model.AssetMetadata{
				Anchor:          model.AnchorConfig{X: 0.5, Y: 0.26},
				ReferenceWidth:  1080,
				ScaleMultiplier: 1.2,
				RotationOffset:  0,
				Offset:          model.OffsetConfig{X: 0, Y: 0},
			},
			Status:    model.StatusPublished,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, it := range items {
		repo.items[it.ID] = it
	}

	return repo
}

func (r *MemoryRepository) FindPublished(ctx context.Context, category string, offset, limit int) ([]model.ClothingItem, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []model.ClothingItem
	for _, item := range r.items {
		if item.Status != model.StatusPublished {
			continue
		}
		if category != "" && item.Category != category {
			continue
		}
		filtered = append(filtered, item)
	}

	total := len(filtered)
	if offset >= total {
		return []model.ClothingItem{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total, nil
}

func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*model.ClothingItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[id]
	if !exists || item.Status != model.StatusPublished {
		return nil, ErrNotFound
	}
	return &item, nil
}

// PostgresRepository provides production PostgreSQL implementation
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) FindPublished(ctx context.Context, category string, offset, limit int) ([]model.ClothingItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM clothing_items WHERE status = $1`
	countArgs := []any{model.StatusPublished}

	if category != "" {
		countQuery += ` AND category = $2`
		countArgs = append(countArgs, category)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count published clothing: %w", err)
	}

	dataQuery := `
		SELECT id, name, description, category, thumbnail_key, asset_key, asset_metadata, status, created_at, updated_at
		FROM clothing_items
		WHERE status = $1
	`
	dataArgs := []any{model.StatusPublished}
	paramIdx := 2

	if category != "" {
		dataQuery += fmt.Sprintf(` AND category = $%d`, paramIdx)
		dataArgs = append(dataArgs, category)
		paramIdx++
	}

	dataQuery += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, paramIdx, paramIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query clothing items: %w", err)
	}
	defer rows.Close()

	var items []model.ClothingItem
	for rows.Next() {
		var item model.ClothingItem
		var metaBytes []byte
		var desc sql.NullString

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&desc,
			&item.Category,
			&item.ThumbnailURL,
			&item.AssetURL,
			&metaBytes,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan clothing item row: %w", err)
		}

		if desc.Valid {
			item.Description = desc.String
		}

		if len(metaBytes) > 0 {
			if err := json.Unmarshal(metaBytes, &item.AssetMetadata); err != nil {
				return nil, 0, fmt.Errorf("failed to parse asset_metadata JSON: %w", err)
			}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	return items, total, nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*model.ClothingItem, error) {
	query := `
		SELECT id, name, description, category, thumbnail_key, asset_key, asset_metadata, status, created_at, updated_at
		FROM clothing_items
		WHERE id = $1 AND status = $2
	`
	var item model.ClothingItem
	var metaBytes []byte
	var desc sql.NullString

	err := r.db.QueryRowContext(ctx, query, id, model.StatusPublished).Scan(
		&item.ID,
		&item.Name,
		&desc,
		&item.Category,
		&item.ThumbnailURL,
		&item.AssetURL,
		&metaBytes,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query item by id: %w", err)
	}

	if desc.Valid {
		item.Description = desc.String
	}

	if len(metaBytes) > 0 {
		if err := json.Unmarshal(metaBytes, &item.AssetMetadata); err != nil {
			return nil, fmt.Errorf("failed to parse asset_metadata JSON: %w", err)
		}
	}

	return &item, nil
}

