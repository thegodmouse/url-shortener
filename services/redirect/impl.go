package redirect

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
	"github.com/thegodmouse/url-shortener/util"
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

// RedirectTo returns the original url of the urlID if exists.
func (s *serviceImpl) RedirectTo(ctx context.Context, id int64) (string, error) {
	shortURL, isCached, err := s.getShortURL(ctx, id)
	if err != nil {
		return "", err
	}
	if !isCached {
		if err := s.cacheStore.Set(ctx, id, shortURL); err != nil {
			log.Errorf("getShortURL: cahce store set err: %v, id: %v", err, id)
		}
	}
	if util.IsRecordExpired(shortURL) || util.IsRecordDeleted(shortURL) {
		return "", util.ErrURLNotFound
	}
	return shortURL.URL, nil
}

func (s *serviceImpl) getShortURL(ctx context.Context, id int64) (*record.ShortURL, bool, error) {
	var err error
	var shortURL *record.ShortURL

	shortURL, err = s.cacheStore.Get(ctx, id)
	if err == nil {
		// cache hit, return the result
		return shortURL, true, nil
	}
	if err != cache.ErrKeyNotFound {
		// suppress error
		log.Errorf("getShortURL: cache store get err: %v, id: %v", err, id)
	}
	shortURL, err = s.dbStore.Get(ctx, id)
	if err != nil {
		return nil, false, err
	}
	return shortURL, false, nil
}
