package browser

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
)

type Opener interface {
	Open(url string) error
}

type OpenerFunc func(url string) error

func (f OpenerFunc) Open(url string) error {
	return f(url)
}

func New() Opener {
	return OpenerFunc(func(url string) error {
		cmd := CommandFor(runtime.GOOS, url)
		if cmd == nil {
			return errors.New("browser open: unsupported platform " + runtime.GOOS)
		}
		return cmd.Start()
	})
}

func CommandFor(goos, url string) *exec.Cmd {
	switch goos {
	case "linux":
		return exec.CommandContext(context.WithoutCancel(context.Background()), "xdg-open", url)
	case "darwin":
		return exec.CommandContext(context.WithoutCancel(context.Background()), "open", url)
	case "windows":
		return exec.CommandContext(context.WithoutCancel(context.Background()), "rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return nil
	}
}
