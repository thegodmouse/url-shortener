package cache

import (
	"context"

	"github.com/thegodmouse/url-shortener/db/record"
)

type Store interface {
	Get(ctx context.Context, id int64) (*record.ShortURL, error)
	Set(ctx context.Context, id int64, record *record.ShortURL) error
}
