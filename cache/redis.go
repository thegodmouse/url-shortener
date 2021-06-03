package cache

import (
	"context"
	"github.com/thegodmouse/url-shortener/db/record"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	defaultExpiration = 10 * time.Minute
)

func NewRedisStore(options *redis.Options) *redisCache {
	return &redisCache{
		client:     redis.NewClient(options),
		expiration: defaultExpiration,
	}
}

type redisCache struct {
	client     *redis.Client
	expiration time.Duration
}

func (r *redisCache) Get(ctx context.Context, urlID string) (*record.ShortURL, error) {
	var record *record.ShortURL
	if err := r.client.Get(ctx, r.makeKey(urlID)).Scan(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (r *redisCache) Set(ctx context.Context, urlID string, record *record.ShortURL) error {
	if err := r.client.Set(ctx, r.makeKey(urlID), record, r.expiration).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Evict(ctx context.Context, urlID string) error {
	if err := r.client.Del(ctx, r.makeKey(urlID)).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisCache) makeKey(urlID string) string {
	return urlID
}
