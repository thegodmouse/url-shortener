package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	// ErrNoRows is an alias for sql.ErrNoRows
	ErrNoRows = sql.ErrNoRows
)

// Store defines the interface for url_shortener database store
type Store interface {
	// Create creates a new short url record or recycles an old one from expired or deleted records.
	Create(ctx context.Context, url string, expireAt time.Time) (*record.ShortURL, error)
	// Get gets the short url record with the given id.
	Get(ctx context.Context, id int64) (*record.ShortURL, error)
	// GetExpiredIDs returns a channel for reading expired ids.
	GetExpiredIDs(ctx context.Context) (<-chan int64, error)
	// Expire expires the short url record with the given id, and makes it recyclable.
	Expire(ctx context.Context, id int64) error
	// Delete deletes the short url record with the given id, and makes is recyclable.
	Delete(ctx context.Context, id int64) error
}
