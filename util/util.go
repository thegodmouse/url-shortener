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
	// ErrURLNotFound is returned when there is no matching url with query.
	ErrURLNotFound = errors.New("short url not found")
)

// IsRecordExpired checks if the given record is expired.
func IsRecordExpired(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.ExpireAt.Before(time.Now())
}

// IsRecordDeleted checks if the given record is deleted.
func IsRecordDeleted(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.IsDeleted
}

// DeleteExpiredURLs is an infinite loop for periodically checking whether there is any expired record in database.
func DeleteExpiredURLs(ctx context.Context, dbStore db.Store, interval time.Duration) <-chan bool {
	log.Infof("DeleteExpiredURLs: check expired records with interval: %v", interval)
	done := make(chan bool, 0)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Infof("DeleteExpiredURLs: received cancel signal, exiting...")
				done <- true
				return
			case <-ticker.C:
				log.Infof("DeleteExpiredURLs: start checking expired record in database")
				ch, err := dbStore.GetExpiredIDs(ctx)
				if err != nil {
					log.Errorf("DeleteExpiredURLs: get expired id err: %v", err)
					continue
				}
				for id := range ch {
					if err := dbStore.Expire(ctx, id); err != nil {
						log.Errorf("DeleteExpiredURLs: expire err: %v, with id: %v", err, id)
					}
				}
			}
		}
	}()
	return done
}
