package spoof_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/laretz/internal/lib/spoof"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func TestMiddleware(t *testing.T) {
	probe := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			return
		}
		_, _ = w.Write([]byte(r.Method + "|" + string(body)))
	})

	t.Run("post with _method=delete reaches probe as delete without the field", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/portfolio/7", "_method=DELETE&id=7")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, "DELETE|id=7", res.Body.String())
	})

	t.Run("plain post stays post with intact body", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/portfolio", "name=main")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, "POST|name=main", res.Body.String())
	})

	t.Run("non-form content type is untouched", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/x", `{"_method":"DELETE"}`)
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, `POST|{"_method":"DELETE"}`, res.Body.String())
	})

	t.Run("other spoof values are ignored", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/x", "_method=PUT&a=1")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, "POST|_method=PUT&a=1", res.Body.String())
	})

	t.Run("malformed form body is passed through", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/x", "%zz_method=DELETE")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, "POST|%zz_method=DELETE", res.Body.String())
	})

	t.Run("delete is case insensitive", func(t *testing.T) {
		req := httpx.NewRequest(t, http.MethodPost, "/portfolio/7", "_method=delete")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		res := httptest.NewRecorder()

		spoof.Middleware(probe).ServeHTTP(res, req)

		assert.Equal(t, "DELETE|", res.Body.String())
	})
}
