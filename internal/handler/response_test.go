package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteSuccess(t *testing.T) {
	t.Run("writes success with data and pagination", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := []string{"item1", "item2"}
		pagination := &PaginationMeta{
			Page:  1,
			Limit: 10,
			Total: 2,
		}

		WriteSuccess(rec, http.StatusOK, data, pagination)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
			t.Errorf("expected Content-Type application/json; charset=utf-8, got %s", contentType)
		}

		var resp SuccessEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Pagination == nil || resp.Pagination.Total != 2 {
			t.Errorf("expected pagination total 2, got %v", resp.Pagination)
		}
	})

	t.Run("writes success with data only and empty pagination", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := map[string]string{"key": "value"}

		WriteSuccess(rec, http.StatusCreated, data, nil)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var resp SuccessEnvelope
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Pagination != nil {
			t.Errorf("expected nil pagination, got %v", resp.Pagination)
		}
	})
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteError(rec, http.StatusBadRequest, "INVALID_INPUT", "Invalid clothing ID", map[string]string{"field": "id"})

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error.Code != "INVALID_INPUT" {
		t.Errorf("expected code INVALID_INPUT, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "Invalid clothing ID" {
		t.Errorf("expected message 'Invalid clothing ID', got %s", resp.Error.Message)
	}
}

