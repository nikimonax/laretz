package plugin

import (
	"context"

	"github.com/go-logr/logr"
)

type Asset struct {
	ID     string
	Symbol string
	Name   string
}

type AssetProvider interface {
	Provide(ctx context.Context) ([]Asset, error)
}

type App interface {
	Logger() logr.Logger

	GetConfig(cfg any) error
	SetConfig(cfg any) error

	RegisterAssetProvider(provider AssetProvider) error
}

type PluginID = string

type Plugin interface {
	ID() PluginID
	Init(app App) error
}
