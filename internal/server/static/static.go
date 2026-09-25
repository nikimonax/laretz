package static

import (
	"embed"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed css js
var files embed.FS

func Mount(r chi.Router) {
	filesServer := http.StripPrefix("/static", http.FileServerFS(files))
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=86400")
		filesServer.ServeHTTP(w, r)
	}))
}
