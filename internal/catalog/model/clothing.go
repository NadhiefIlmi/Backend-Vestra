package model

import "time"

type AnchorConfig struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type OffsetConfig struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type AssetMetadata struct {
	Anchor          AnchorConfig `json:"anchor"`
	ReferenceWidth  int          `json:"referenceWidth"`
	ScaleMultiplier float64      `json:"scaleMultiplier"`
	RotationOffset  float64      `json:"rotationOffset"`
	Offset          OffsetConfig `json:"offset"`
}

type ClothingItem struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	Category      string        `json:"category"`
	ThumbnailURL  string        `json:"thumbnailUrl"`
	AssetURL      string        `json:"assetUrl"`
	AssetMetadata AssetMetadata `json:"assetMetadata"`
	Status        string        `json:"status"` // draft, published, archived
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

