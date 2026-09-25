package middleware

import (
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

func RequestLog(logger logr.Logger) func(http.Handler) http.Handler {
	return chimw.RequestLogger(requestLogFormatter{logger: logger})
}

type requestLogFormatter struct {
	logger logr.Logger
}

func (f requestLogFormatter) NewLogEntry(r *http.Request) chimw.LogEntry {
	return &requestLogEntry{logger: f.logger, method: r.Method, path: r.URL.Path}
}

type requestLogEntry struct {
	logger logr.Logger
	method string
	path   string
}

func (e *requestLogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	e.logger.Info("request", "method", e.method, "path", e.path, "status", status, "bytes", bytes, "duration", elapsed)
}

func (e *requestLogEntry) Panic(value any, stack []byte) {
	e.logger.Error(nil, "request panic", "value", value, "stack", string(stack))
}
