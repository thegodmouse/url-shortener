package redirect

import (
	"context"
	"database/sql"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
	"github.com/thegodmouse/url-shortener/util"
)

// NewService returns a new redirect.Service with default implementation.
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

// RedirectTo returns the original url with given id.
func (s *serviceImpl) RedirectTo(ctx context.Context, id int64) (string, error) {
	shortURL, isCached, err := s.getShortURL(ctx, id)
	if err != nil {
		log.Errorf("redirect.RedirectTo: get short url record err: %v, with id: %v", err, id)
		if err == sql.ErrNoRows {
			if err := s.cacheStore.Set(ctx, id, &record.ShortURL{ID: id, IsNotExist: true}); err != nil {
				log.Errorf("redirect.RedirectTo: cahce store set err: %v, with id: %v", err, id)
			}
		}
		return "", err
	}
	if !isCached {
		log.Infof("redirect.RedirectTo: not found in cache, try to set cache with id: %v", id)
		if err := s.cacheStore.Set(ctx, id, shortURL); err != nil {
			log.Errorf("redirect.RedirectTo: cahce store set err: %v, with id: %v", err, id)
		}
	}
	// check if the record is expired or deleted
	if util.IsRecordExpired(shortURL) || util.IsRecordDeleted(shortURL) || util.IsRecordNotExist(shortURL) {
		log.Errorf("redirect.RedirectTo: short url is unavailable, url record: %+v", shortURL)
		return "", util.ErrURLNotFound
	}
	log.Infof("redirect.RedirectTo: successfully get the original url from the record: %v, with id: %v", shortURL, id)
	return shortURL.URL, nil
}

func (s *serviceImpl) getShortURL(ctx context.Context, id int64) (*record.ShortURL, bool, error) {
	shortURL, err := s.cacheStore.Get(ctx, id)
	if err == nil {
		// cache hit, return the result
		return shortURL, true, nil
	}
	if err != cache.ErrKeyNotFound {
		// suppress error
		log.Errorf("redirect.getShortURL: cache store get err: %v, id: %v", err, id)
	}
	shortURL, err = s.dbStore.Get(ctx, id)
	if err != nil {
		return nil, false, err
	}
	return shortURL, false, nil
}
