package browser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nikimonax/laretz/internal/lib/browser"
)

func TestCommandForPlatform(t *testing.T) {
	tests := []struct {
		goos     string
		wantPath string
		wantArgs []string
		wantNil  bool
	}{
		{goos: "linux", wantPath: "xdg-open", wantArgs: []string{"http://127.0.0.1:8080"}},
		{goos: "darwin", wantPath: "open", wantArgs: []string{"http://127.0.0.1:8080"}},
		{goos: "windows", wantPath: "rundll32", wantArgs: []string{"url.dll,FileProtocolHandler", "http://127.0.0.1:8080"}},
		{goos: "plan9", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			cmd := browser.CommandFor(tt.goos, "http://127.0.0.1:8080")

			if tt.wantNil {
				assert.Nil(t, cmd)
				return
			}
			assert.Equal(t, tt.wantPath, cmd.Args[0])
			assert.Equal(t, append([]string{tt.wantPath}, tt.wantArgs...), cmd.Args)
			assert.True(t, strings.HasSuffix(cmd.Path, tt.wantPath), "resolved path must point at %s", tt.wantPath)
		})
	}
}

func TestOpenerFuncIsASeam(t *testing.T) {
	var opened []string
	var o browser.Opener = browser.OpenerFunc(func(url string) error {
		opened = append(opened, url)
		return nil
	})

	assert.NoError(t, o.Open("http://127.0.0.1:9000"))
	assert.NoError(t, o.Open("http://127.0.0.1:9001"))
	assert.Equal(t, []string{"http://127.0.0.1:9000", "http://127.0.0.1:9001"}, opened)
}

func TestOpenerErrorIsWrappedNotPanicking(t *testing.T) {
	var o browser.Opener = browser.OpenerFunc(func(string) error { return errors.New("boom") })
	assert.EqualError(t, o.Open("http://x"), "boom")
}
