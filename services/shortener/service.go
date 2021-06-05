package shortener

import (
	"context"
	"time"
)

type Service interface {
	Shorten(ctx context.Context, url string, expireAt time.Time) (int64, error)
	Delete(ctx context.Context, id int64) error
}
