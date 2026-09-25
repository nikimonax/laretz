package spoof

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	methodField  = "_method"
	maxFormBytes = 1 << 20
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && isURLEncoded(r) {
			r = rewrite(w, r)
		}
		next.ServeHTTP(w, r)
	})
}

func isURLEncoded(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
}

func rewrite(w http.ResponseWriter, r *http.Request) *http.Request {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxFormBytes))
	_ = r.Body.Close()
	if err != nil {
		r.Body = http.NoBody
		return r
	}

	values, parseErr := url.ParseQuery(string(body))
	if parseErr != nil || !strings.EqualFold(values.Get(methodField), http.MethodDelete) {
		r.Body = io.NopCloser(bytes.NewReader(body))
		return r
	}

	values.Del(methodField)
	encoded := values.Encode()
	spoofed := r.Clone(r.Context())
	spoofed.Method = http.MethodDelete
	spoofed.Body = io.NopCloser(strings.NewReader(encoded))
	spoofed.ContentLength = int64(len(encoded))
	spoofed.Header.Set("Content-Length", strconv.Itoa(len(encoded)))
	return spoofed
}
