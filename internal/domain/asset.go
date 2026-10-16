package domain

import (
	"time"
)

type ExternalAssetID = int64
type SourceID = string

type ExternalAsset struct {
	ID         ExternalAssetID
	SourceID   SourceID
	ExternalID string
	Symbol     string
	Name       string
	SyncedAt   time.Time
}

type AssetID = int64

type Asset struct {
	ID        AssetID
	Symbol    string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time

	External ExternalAsset
}
