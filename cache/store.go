package cache

import (
	"context"

	"github.com/thegodmouse/url-shortener/db/record"
)

type Store interface {
	Get(ctx context.Context, urlID string) (*record.ShortURL, error)
	Set(ctx context.Context, urlID string, record *record.ShortURL) error
	Evict(ctx context.Context, urlID string) error
}
