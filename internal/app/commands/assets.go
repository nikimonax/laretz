package commands

import "github.com/nikimonax/laretz/internal/app/dto"

type SyncExternalAssetsCommand struct {
	Assets []dto.ExternalAsset
}
