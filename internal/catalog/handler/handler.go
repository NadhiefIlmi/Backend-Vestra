package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/live-camera-ar-tryon/backend/internal/catalog/repository"
	"github.com/live-camera-ar-tryon/backend/internal/catalog/service"
	"github.com/live-camera-ar-tryon/backend/internal/handler"
)

type CatalogHandler struct {
	svc *service.Service
}

func New(svc *service.Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) ListClothing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handler.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", nil)
		return
	}

	query := r.URL.Query()
	category := strings.TrimSpace(query.Get("category"))

	page := 1
	if pStr := query.Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		} else {
			handler.WriteError(w, http.StatusBadRequest, "INVALID_PARAMETER", "Parameter 'page' harus berupa angka positif", nil)
			return
		}
	}

	limit := 20
	if lStr := query.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		} else {
			handler.WriteError(w, http.StatusBadRequest, "INVALID_PARAMETER", "Parameter 'limit' harus berupa angka positif", nil)
			return
		}
	}

	result, err := h.svc.ListClothing(r.Context(), service.ListClothingParams{
		Category: category,
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar katalog", nil)
		return
	}

	handler.WriteSuccess(w, http.StatusOK, result.Items, &handler.PaginationMeta{
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	})
}

func (h *CatalogHandler) GetClothingByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handler.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", nil)
		return
	}

	// Extract {id} from path: /api/v1/clothing/{id}
	prefix := "/api/v1/clothing/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		handler.WriteError(w, http.StatusBadRequest, "INVALID_PATH", "Path tidak valid", nil)
		return
	}

	id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, prefix))
	if id == "" || strings.Contains(id, "/") {
		handler.WriteError(w, http.StatusBadRequest, "INVALID_ID", "ID pakaian wajib diisi", nil)
		return
	}

	item, err := h.svc.GetClothingByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			handler.WriteError(w, http.StatusNotFound, "NOT_FOUND", "Item pakaian tidak ditemukan", nil)
			return
		}
		handler.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil detail item", nil)
		return
	}

	handler.WriteSuccess(w, http.StatusOK, item, nil)
}

