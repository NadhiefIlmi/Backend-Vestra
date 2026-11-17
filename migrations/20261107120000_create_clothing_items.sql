-- Migration: Create clothing_items table
-- Up
CREATE TABLE IF NOT EXISTS clothing_items (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(64) NOT NULL DEFAULT 'top',
    thumbnail_key TEXT NOT NULL,
    asset_key TEXT NOT NULL,
    asset_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient catalog querying
CREATE INDEX IF NOT EXISTS idx_clothing_items_status ON clothing_items(status);
CREATE INDEX IF NOT EXISTS idx_clothing_items_category_status ON clothing_items(category, status);
CREATE INDEX IF NOT EXISTS idx_clothing_items_created_at ON clothing_items(created_at DESC);

-- Seed initial catalog items for virtual try-on
INSERT INTO clothing_items (id, name, description, category, thumbnail_key, asset_key, asset_metadata, status)
VALUES
(
    'shirt-basic-black',
    'Classic Black T-Shirt',
    'Kaos hitam katun combed 30s premium dengan fitting regular modern.',
    'top',
    '/assets/clothing/shirt-black-thumb.webp',
    '/assets/clothing/shirt-black.png',
    '{"anchor": {"x": 0.5, "y": 0.16}, "referenceWidth": 1000, "scaleMultiplier": 1.0, "rotationOffset": 0, "offset": {"x": 0, "y": 0}}'::jsonb,
    'published'
),
(
    'shirt-basic-white',
    'Classic White T-Shirt',
    'Kaos putih netral dengan kerah bulat lembut, pas untuk daily wear.',
    'top',
    '/assets/clothing/shirt-white-thumb.webp',
    '/assets/clothing/shirt-white.png',
    '{"anchor": {"x": 0.5, "y": 0.16}, "referenceWidth": 1000, "scaleMultiplier": 1.0, "rotationOffset": 0, "offset": {"x": 0, "y": 0}}'::jsonb,
    'published'
),
(
    'hoodie-urban-grey',
    'Urban Oversized Hoodie',
    'Hoodie abu-abu streetwear fleece tebal dengan potongan santai.',
    'top',
    '/assets/clothing/hoodie-grey-thumb.webp',
    '/assets/clothing/hoodie-grey.png',
    '{"anchor": {"x": 0.5, "y": 0.14}, "referenceWidth": 1080, "scaleMultiplier": 1.08, "rotationOffset": 0, "offset": {"x": 0, "y": 0}}'::jsonb,
    'published'
)
ON CONFLICT (id) DO NOTHING;

