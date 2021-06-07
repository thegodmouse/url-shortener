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
	GetExpiredIDs(ctx context.Context) (<-chan int64, error)
	Expire(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}
