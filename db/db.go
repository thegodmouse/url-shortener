package db

import (
	"context"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

type Store interface {
	Create(ctx context.Context, url string, expireAt time.Time) (*record.ShortURL, error)
	Get(ctx context.Context, id int64) (*record.ShortURL, error)
	Delete(ctx context.Context, id int64) error
}
