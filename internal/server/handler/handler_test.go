package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/server/handler"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func TestHome(t *testing.T) {
	res := httptest.NewRecorder()
	handler.New(logr.Discard()).ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, "/"))

	require.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, res.Header().Get("Content-Type"), "text/html")

	body := res.Body.String()
	assert.Contains(t, body, "Laretz")
	assert.Contains(t, body, `href="/settings"`)
	assert.Contains(t, body, "/static/css/pico.min.css")
}

func TestRouterSmoke(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		origin     string
		wantStatus int
	}{
		{name: "home is rendered", method: http.MethodGet, path: "/", wantStatus: http.StatusOK},
		{name: "static asset is served through the router", method: http.MethodGet, path: "/static/css/pico.min.css", wantStatus: http.StatusOK},
		{name: "unknown page is 404", method: http.MethodGet, path: "/nope", wantStatus: http.StatusNotFound},
		{name: "foreign origin post is forbidden", method: http.MethodPost, path: "/portfolio", origin: "http://evil.example", wantStatus: http.StatusForbidden},
		{name: "spoofed delete with matching origin passes origin check", method: http.MethodPost, path: "/portfolio", body: "_method=DELETE", origin: "http://127.0.0.1:8080", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httpx.NewRequest(t, tt.method, tt.path, tt.body)
			req.Host = "127.0.0.1:8080"
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			res := httptest.NewRecorder()
			handler.New(logr.Discard()).ServeHTTP(res, req)

			assert.Equal(t, tt.wantStatus, res.Code)
		})
	}
}
