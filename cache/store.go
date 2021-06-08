package cache

import (
	"context"

	"github.com/thegodmouse/url-shortener/db/record"
)

// Store defines the interface for url_shortener cache store.
type Store interface {
	// Get gets the record with id from the cache.
	Get(ctx context.Context, id int64) (*record.ShortURL, error)
	// Set sets the record with id to the cache.
	Set(ctx context.Context, id int64, record *record.ShortURL) error
}
