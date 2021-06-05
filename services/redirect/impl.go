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
	shortURL, err := s.getShortURL(ctx, id)
	if err != nil {
		return "", err
	}
	if util.IsRecordExpired(shortURL) {
		return "", util.ErrURLExpired
	}
	return shortURL.URL, nil
}

func (s *serviceImpl) getShortURL(ctx context.Context, id int64) (*record.ShortURL, error) {
	var err error
	var shortURL *record.ShortURL

	shortURL, err = s.cacheStore.Get(ctx, id)
	if err == nil {
		// cache hit, return the result
		return shortURL, nil
	}
	if err != cache.ErrKeyNotFound {
		// suppress error
		log.Errorf("getShortURL: cache store err: %v, with id: %v", err, id)
	}
	shortURL, err = s.dbStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return shortURL, nil
}
