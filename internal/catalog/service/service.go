package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/model"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/repository"
)

var (
	ErrInvalidPagination = errors.New("invalid page or limit parameters")
	ErrInvalidID         = errors.New("invalid clothing id")
)

type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

type ListClothingParams struct {
	Category string
	Page     int
	Limit    int
}

type ListClothingResult struct {
	Items []model.ClothingItem
	Page  int
	Limit int
	Total int
}

func (s *Service) ListClothing(ctx context.Context, params ListClothingParams) (*ListClothingResult, error) {
	page := params.Page
	if page < 1 {
		page = 1
	}

	limit := params.Limit
	if limit < 1 {
		limit = 20
	} else if limit > 100 {
		limit = 100 // Cap max items per page
	}

	offset := (page - 1) * limit

	items, total, err := s.repo.FindPublished(ctx, params.Category, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch clothing list: %w", err)
	}

	return &ListClothingResult{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *Service) GetClothingByID(ctx context.Context, id string) (*model.ClothingItem, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get clothing item: %w", err)
	}

	return item, nil
}

