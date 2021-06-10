package shortener

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
	"github.com/thegodmouse/url-shortener/util"
)

// NewService returns a new shorten.Service with default implementation.
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

// Shorten shortens an url with an unique id, and create a record in the database.
func (s *serviceImpl) Shorten(ctx context.Context, url string, expireAt time.Time) (int64, error) {
	// create short url record in database
	shortURL, err := s.dbStore.Create(ctx, url, expireAt)
	if err != nil {
		return 0, err
	}
	// set short url record in database for further redirect queries.
	if err := s.cacheStore.Set(ctx, shortURL.ID, shortURL); err != nil {
		log.Errorf("shortener.Shorten: cache store set err: %v, id: %v", err, shortURL.ID)
	}
	log.Infof("shortener.Shorten: finished shorten url with id: %v", shortURL.ID)
	return shortURL.ID, nil
}

// Delete deletes an url with id.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	// lookup id in the cache to see if this record is deleted.
	shortURL, err := s.cacheStore.Get(ctx, id)
	if err != nil {
		log.Errorf("shortener.Delete: cache store get err: %v, with id: %v", err, id)
	} else if util.IsRecordDeleted(shortURL) || util.IsRecordNotExist(shortURL) {
		// found in the cache, and the record is deleted or not exist.
		log.Infof("shortener.Delete: record is not available for id: %v", id)
		return nil
	}
	// the record is either not exist in the cache or not deleted, directly delete it in the database.
	if err := s.dbStore.Delete(ctx, id); err != nil {
		log.Errorf("shortener.Delete: db store delete err: %v, with id: %v", err, id)
		return err
	}
	// set record as deleted in the cache for the next call
	if err := s.cacheStore.Set(ctx, id, &record.ShortURL{ID: id, IsDeleted: true}); err != nil {
		log.Errorf("shortener.Delete: cache store set err: %v, with id: %v", err, id)
	}
	log.Infof("shortener.Delete: finished deleting record with id: %v", id)
	return nil
}
