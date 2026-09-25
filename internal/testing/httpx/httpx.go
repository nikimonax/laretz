package httpx

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewRequest(t testing.TB, method, target string, body ...string) *http.Request {
	t.Helper()
	var reader io.Reader
	if len(body) > 0 && body[0] != "" {
		reader = strings.NewReader(body[0])
	}
	req, err := http.NewRequestWithContext(context.Background(), method, target, reader)
	require.NoError(t, err)
	return req
}
