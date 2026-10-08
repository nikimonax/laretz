package httplog

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type LoggingRoundTripper struct {
	next   http.RoundTripper
	logger logr.Logger
}

// RoundTrip implements [http.RoundTripper].
func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := rt.next.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		rt.logger.Error(err, "request failed",
			"method", req.Method,
			"url", req.URL.String(),
			"duration", duration,
		)
		return nil, err
	}

	rt.logger.Info("request",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.Status,
		"duration", duration,
	)
	return resp, nil
}

func NewLoggingRoundTripper(next http.RoundTripper, logger logr.Logger) http.RoundTripper {
	return &LoggingRoundTripper{next: next, logger: logger}
}
