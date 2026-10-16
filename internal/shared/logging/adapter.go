package logging

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewFxLogger(logger *zap.Logger) fxevent.Logger {
	fxLogger := &fxevent.ZapLogger{Logger: logger}
	fxLogger.UseLogLevel(zapcore.DebugLevel)
	return fxLogger
}

func NewLoggingMiddleware(logger logr.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				logger.Info(
					"handle request",
					"method", r.Method,
					"uri", r.RequestURI,
					"status", ww.Status(),
					"duration", time.Since(t1),
					"size", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

type GooseLoggerAdapter struct {
	sugar *zap.SugaredLogger
}

// Fatalf implements [goose.Logger].
func (adapter *GooseLoggerAdapter) Fatalf(format string, v ...any) {
	adapter.sugar.Fatalf(format, v...)
}

// Printf implements [goose.Logger].
func (adapter *GooseLoggerAdapter) Printf(format string, v ...any) {
	adapter.sugar.Infof(format, v...)
}

func NewGooseLoggerAdapter(sugar *zap.SugaredLogger) goose.Logger {
	return &GooseLoggerAdapter{sugar: sugar}
}
