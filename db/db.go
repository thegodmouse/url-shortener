package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	ErrNoRows = sql.ErrNoRows
)

type Store interface {
	Create(ctx context.Context, url string, expireAt time.Time) (*record.ShortURL, error)
	Get(ctx context.Context, id int64) (*record.ShortURL, error)
	Delete(ctx context.Context, id int64) error
}
