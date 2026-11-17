package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/model"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/repository"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/service"
	baseHandler "github.com/live-camera-ar-tryon/backend/internal/handler"
)

func setupTestHandler() *CatalogHandler {
	repo := repository.NewMemoryRepository()
	svc := service.New(repo)
	return New(svc)
}

func TestListClothingHandler(t *testing.T) {
	h := setupTestHandler()

	// 1. Success test
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clothing?page=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.ListClothing(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Data       []model.ClothingItem        `json:"data"`
		Pagination *baseHandler.PaginationMeta `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if len(resp.Data) == 0 {
		t.Errorf("expected items in response data, got 0")
	}
	if resp.Pagination == nil || resp.Pagination.Total == 0 {
		t.Errorf("expected valid pagination metadata")
	}

	// 2. Invalid page parameter test
	invalidReq := httptest.NewRequest(http.MethodGet, "/api/v1/clothing?page=abc", nil)
	invalidRec := httptest.NewRecorder()

	h.ListClothing(invalidRec, invalidReq)

	if invalidRec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid page param, got %d", invalidRec.Code)
	}
}

func TestGetClothingByIDHandler(t *testing.T) {
	h := setupTestHandler()

	// 1. Success test
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clothing/shirt-basic-black", nil)
	rec := httptest.NewRecorder()

	h.GetClothingByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Data model.ClothingItem `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}
	if resp.Data.ID != "shirt-basic-black" {
		t.Errorf("expected id 'shirt-basic-black', got '%s'", resp.Data.ID)
	}

	// 2. 404 Not Found test
	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/v1/clothing/non-existent-item", nil)
	notFoundRec := httptest.NewRecorder()

	h.GetClothingByID(notFoundRec, notFoundReq)

	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", notFoundRec.Code)
	}
}

