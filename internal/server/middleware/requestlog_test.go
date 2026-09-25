package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/go-logr/zapr"

	"github.com/nikimonax/laretz/internal/server/middleware"
	"github.com/nikimonax/laretz/internal/testing/httpx"
)

func TestRequestLog(t *testing.T) {
	core, records := observer.New(zap.DebugLevel)
	logger := zapr.NewLogger(zap.New(core))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("body"))
	})

	res := httptest.NewRecorder()
	middleware.RequestLog(logger)(next).ServeHTTP(res, httpx.NewRequest(t, http.MethodPost, "/portfolio/7"))

	assert.Equal(t, http.StatusTeapot, res.Code)

	entries := records.All()
	require.Len(t, entries, 1)
	entry := entries[0]
	assert.Equal(t, zapcore.InfoLevel, entry.Level)
	assert.Equal(t, "request", entry.Message)

	fields := map[string]zapcore.Field{}
	for _, f := range entry.Context {
		fields[f.Key] = f
	}
	assert.Equal(t, "POST", fields["method"].String)
	assert.Equal(t, "/portfolio/7", fields["path"].String)
	assert.Equal(t, int64(http.StatusTeapot), fields["status"].Integer)
	assert.Positive(t, fields["duration"].Integer, "duration must be positive")
}

func TestRequestLogDefaultsTo200(t *testing.T) {
	core, records := observer.New(zap.DebugLevel)
	logger := zapr.NewLogger(zap.New(core))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("no explicit status"))
	})

	res := httptest.NewRecorder()
	middleware.RequestLog(logger)(next).ServeHTTP(res, httpx.NewRequest(t, http.MethodGet, "/"))

	entries := records.All()
	require.Len(t, entries, 1)
	for _, f := range entries[0].Context {
		if f.Key == "status" {
			assert.Equal(t, int64(http.StatusOK), f.Integer)
		}
	}
}
