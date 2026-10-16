package app

import "errors"

var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrAssetInUse     = errors.New("asset in use")
	ErrSourceMismatch = errors.New("source mismatch")
)
