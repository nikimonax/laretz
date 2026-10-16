package config

import (
	"errors"
	"fmt"
)

var (
	ErrConfig = errors.New("err config")
)

func NewErrConfig(msg string) error {
	return fmt.Errorf("%w: %s", ErrConfig, msg)
}
