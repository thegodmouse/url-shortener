package util

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	ErrURLNotFound = errors.New("short url not found")
)

func IsRecordExpired(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.ExpireAt.Before(time.Now())
}

func IsRecordDeleted(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.IsDeleted
}

func DeleteExpiredURLs(ctx context.Context, dbStore db.Store, interval time.Duration) <-chan bool {
	done := make(chan bool, 0)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				done <- true
				break
			case <-ticker.C:
				ch, err := dbStore.GetExpiredIDs(ctx)
				if err != nil {
					log.Errorf("DeleteExpiredURLs: get expired id err: %v", err)
					continue
				}
				for id := range ch {
					if err := dbStore.Expire(ctx, id); err != nil {
						log.Errorf("DeleteExpiredURLs: expire err: %v, id: %v", err, id)
					}
				}
			}
		}
	}()
	return done
}
