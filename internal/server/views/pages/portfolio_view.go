package pages

import (
	"github.com/nikimonax/laretz/internal/domain"
)

func assetSymbol(assets []domain.Asset, id domain.AssetID) string {
	for _, asset := range assets {
		if asset.ID == id {
			return asset.Symbol
		}
	}

	return ""
}
