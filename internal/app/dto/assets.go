package dto

import "github.com/nikimonax/laretz/internal/domain"

type ExternalAsset struct {
	SourceID   domain.SourceID
	ExternalID string
	Symbol     string
	Name       string
}
