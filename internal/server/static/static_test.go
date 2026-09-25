package static_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikimonax/laretz/internal/server/static"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func mountedRouter() *chi.Mux {
	r := chi.NewRouter()
	static.Mount(r)
	return r
}

func TestStaticServesEmbeddedAssets(t *testing.T) {
	router := mountedRouter()

	tests := []struct {
		path        string
		wantType    string
		wantContent string
	}{
		{path: "/static/css/pico.min.css", wantType: "text/css", wantContent: "pico"},
		{path: "/static/css/uplot.min.css", wantType: "text/css", wantContent: ".uplot"},
		{path: "/static/js/uPlot.iife.min.js", wantType: "text/javascript", wantContent: "uPlot"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			router.ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, tt.path))

			require.Equal(t, http.StatusOK, res.Code)
			assert.NotEmpty(t, res.Body.Bytes(), "body must be non-empty")
			assert.Contains(t, res.Header().Get("Content-Type"), tt.wantType)
			assert.Contains(t, strings.ToLower(res.Body.String()), strings.ToLower(tt.wantContent))
			assert.Equal(t, "public, max-age=86400", res.Header().Get("Cache-Control"))
		})
	}
}

func TestStaticRejectsUnknownAndListings(t *testing.T) {
	router := mountedRouter()

	t.Run("missing file is 404", func(t *testing.T) {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, "/static/css/nope.css"))
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	t.Run("directory listing is 404", func(t *testing.T) {
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, "/static/"))
		assert.Equal(t, http.StatusNotFound, res.Code)
	})
}

func TestEmbeddedFSHasNoCDNPlaceholders(t *testing.T) {
	router := mountedRouter()
	res := httptest.NewRecorder()
	router.ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, "/static/css/pico.min.css"))

	assert.NotContains(t, res.Body.String(), "cdn.jsdelivr.net")
}
