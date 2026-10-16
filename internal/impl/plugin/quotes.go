package plugin

import (
	"errors"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

var (
	ErrQuoteFeedNotFound          = errors.New("quote feed not found")
	ErrQuoteFeedAlreadyRegistered = errors.New("quote feed already registered")
)

type QuoteFeedRegistry struct {
	feeds map[domain.SourceID]interfaces.QuoteFeed
}

// Find implements [interfaces.QuoteFeedRegistry].
func (registry *QuoteFeedRegistry) Find(sourceID domain.SourceID) (interfaces.QuoteFeed, error) {
	feed, ok := registry.feeds[sourceID]

	if !ok {
		return nil, ErrQuoteFeedNotFound
	}

	return feed, nil
}

func (registry *QuoteFeedRegistry) Register(sourceID domain.SourceID, feed interfaces.QuoteFeed) error {
	_, exists := registry.feeds[sourceID]

	if exists {
		return ErrQuoteFeedAlreadyRegistered
	}

	registry.feeds[sourceID] = feed
	return nil
}

var _ interfaces.QuoteFeedRegistry = (*QuoteFeedRegistry)(nil)

func NewQuoteFeedRegistry() *QuoteFeedRegistry {
	return &QuoteFeedRegistry{feeds: make(map[domain.SourceID]interfaces.QuoteFeed)}
}
