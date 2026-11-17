package decart

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/live-camera-ar-tryon/backend/internal/handler"
)

type TokenHandler struct {
	apiKey     string
	httpClient *http.Client
}

func NewTokenHandler(apiKey string) *TokenHandler {
	return &TokenHandler{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *TokenHandler) CreateClientToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handler.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method must be POST", nil)
		return
	}

	if h.apiKey == "" {
		handler.WriteError(w, http.StatusInternalServerError, "CONFIG_ERROR", "DECART_API_KEY is not configured", nil)
		return
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"expiresIn": 600,
	})

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://api.decart.ai/v1/client/tokens", bytes.NewBuffer(reqBody))
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "REQUEST_BUILD_ERROR", err.Error(), nil)
		return
	}

	req.Header.Set("X-API-KEY", h.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		handler.WriteError(w, http.StatusBadGateway, "DECART_API_ERROR", "Failed to connect to Decart API: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		handler.WriteError(w, http.StatusInternalServerError, "READ_ERROR", "Failed to read Decart response", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)
}

