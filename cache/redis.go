package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	ErrKeyNotFound = redis.Nil
)

const (
	defaultExpiration = 10 * time.Minute
)

func NewRedisStore(addr string, password string) *redisCache {
	return &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0, // use default DB
		}),
		expiration: defaultExpiration,
	}
}

type redisCache struct {
	client     *redis.Client
	expiration time.Duration
}

func (r *redisCache) Get(ctx context.Context, id int64) (*record.ShortURL, error) {
	var shortURL *record.ShortURL
	if err := r.client.Get(ctx, r.makeKey(id)).Scan(shortURL); err != nil {
		return nil, err
	}
	return shortURL, nil
}

func (r *redisCache) Set(ctx context.Context, id int64, record *record.ShortURL) error {
	if err := r.client.Set(ctx, r.makeKey(id), record, r.expiration).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Evict(ctx context.Context, id int64) error {
	if err := r.client.Del(ctx, r.makeKey(id)).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) makeKey(id int64) string {
	return fmt.Sprintf("id#%v", id)
}
