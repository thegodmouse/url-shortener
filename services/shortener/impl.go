package shortener

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
)

func NewService(dbStore db.Store, cacheStore cache.Store) *serviceImpl {
	return &serviceImpl{
		dbStore:    dbStore,
		cacheStore: cacheStore,
	}
}

type serviceImpl struct {
	dbStore    db.Store
	cacheStore cache.Store
}

func (s *serviceImpl) Shorten(ctx context.Context, url string, expireAt time.Time) (int64, error) {
	var err error
	var shortURL *record.ShortURL

	shortURL, err = s.dbStore.Create(ctx, url, expireAt)
	if err != nil {
		return 0, err
	}
	err = s.cacheStore.Set(ctx, shortURL.ID, shortURL)
	if err != nil {
		log.Errorf("Shorten: set cache err: %v, urlID: %v", err, shortURL.ID)
	}
	return shortURL.ID, nil
}

func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	if err := s.dbStore.Delete(ctx, id); err != nil {
		return err
	}
	if err := s.cacheStore.Evict(ctx, id); err != nil {
		log.Errorf("Delete: evict cache err: %v, id: %v", err, id)
	}
	return nil
}