package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	// ErrKeyNotFound is an alias for redis.Nil
	ErrKeyNotFound = redis.Nil
)

const (
	defaultExpiration = 10 * time.Minute
)

// NewRedisStore returns a new cache.Store which is implemented by redis cache.
func NewRedisStore(addr string, password string) *redisCache {
	return newRedisStore(
		redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0, // use default DB
		}))
}

func newRedisStore(redisClient *redis.Client) *redisCache {
	return &redisCache{
		client:     redisClient,
		expiration: defaultExpiration,
	}
}

type redisCache struct {
	client     *redis.Client
	expiration time.Duration
}

// Get gets the record with id from the cache.
func (r *redisCache) Get(ctx context.Context, id int64) (*record.ShortURL, error) {
	shortURL := &record.ShortURL{}
	if err := r.client.Get(ctx, r.makeKey(id)).Scan(shortURL); err != nil {
		if err != redis.Nil {
			log.Errorf("redisCache.Get: get from cache err: %v, id: %v", err, id)
		}
		return nil, err
	}
	return shortURL, nil
}

// Set sets the record with id to the cache.
func (r *redisCache) Set(ctx context.Context, id int64, record *record.ShortURL) error {
	if err := r.client.Set(ctx, r.makeKey(id), record, r.expiration).Err(); err != nil {
		log.Errorf("redisCache.Set: set cache err: %v, id: %v, data: %+v", err, id, record)
		return err
	}
	return nil
}

func (r *redisCache) makeKey(id int64) string {
	return fmt.Sprintf("id#%v", id)
}
