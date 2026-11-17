package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value(RequestIDKey)
		if reqID == nil || reqID == "" {
			t.Errorf("expected request id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	respReqID := rec.Header().Get("X-Request-ID")
	if respReqID == "" {
		t.Errorf("expected X-Request-ID response header")
	}
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	panickingHandler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected critical error")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	panickingHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 on panic recovery, got %d", rec.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	cors := CORS([]string{"http://localhost:5173"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Allowed origin
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	cors.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("expected Access-Control-Allow-Origin header set for allowed origin")
	}

	// 2. Preflight OPTIONS
	optionsReq := httptest.NewRequest(http.MethodOptions, "/test", nil)
	optionsReq.Header.Set("Origin", "http://localhost:5173")
	optionsRec := httptest.NewRecorder()

	cors.ServeHTTP(optionsRec, optionsReq)

	if optionsRec.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS request, got %d", optionsRec.Code)
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'",
	}

	for header, expectedVal := range expectedHeaders {
		if val := rec.Header().Get(header); val != expectedVal {
			t.Errorf("expected header %s to be '%s', got '%s'", header, expectedVal, val)
		}
	}
}

func TestBodyLimitMiddleware(t *testing.T) {
	limitHandler := BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 20)
		_, err := r.Body.Read(buf)
		if err != nil {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Payload with 15 bytes exceeds 10 bytes limit
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("123456789012345"))
	rec := httptest.NewRecorder()

	limitHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413 Payload Too Large, got %d", rec.Code)
	}
}

