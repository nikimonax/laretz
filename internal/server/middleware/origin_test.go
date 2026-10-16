package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/laretz/internal/server/middleware"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestOriginCheck(t *testing.T) {
	handler := middleware.OriginCheck(http.HandlerFunc(ok))

	tests := []struct {
		name       string
		method     string
		origin     string
		host       string
		wantStatus int
	}{
		{name: "get is exempt even with foreign origin", method: http.MethodGet, origin: "http://evil.example", host: "127.0.0.1:8080", wantStatus: http.StatusOK},
		{name: "missing origin is allowed", method: http.MethodPost, origin: "", host: "127.0.0.1:8080", wantStatus: http.StatusOK},
		{name: "matching host and port is allowed", method: http.MethodPost, origin: "http://127.0.0.1:8080", host: "127.0.0.1:8080", wantStatus: http.StatusOK},
		{name: "matching loopback name is allowed", method: http.MethodPost, origin: "http://localhost:8080", host: "localhost:8080", wantStatus: http.StatusOK},
		{name: "loopback variants with different ports are allowed", method: http.MethodPost, origin: "http://localhost:8080", host: "127.0.0.1:9090", wantStatus: http.StatusOK},
		{name: "foreign origin is rejected", method: http.MethodPost, origin: "http://evil.example", host: "127.0.0.1:8080", wantStatus: http.StatusForbidden},
		{name: "foreign origin with matching port is rejected", method: http.MethodPost, origin: "http://evil.example:8080", host: "127.0.0.1:8080", wantStatus: http.StatusForbidden},
		{name: "opaque origin is rejected", method: http.MethodPost, origin: "null", host: "127.0.0.1:8080", wantStatus: http.StatusForbidden},
		{name: "delete is checked too", method: http.MethodDelete, origin: "http://evil.example", host: "127.0.0.1:8080", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httpx.NewRequest(t, tt.method, "/")
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			req.Host = tt.host
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			assert.Equal(t, tt.wantStatus, res.Code)
		})
	}
}
