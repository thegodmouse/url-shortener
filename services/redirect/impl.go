package redirect

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/converter"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
	"github.com/thegodmouse/url-shortener/util"
)

func NewService(dbStore db.Store, cacheStore cache.Store, conv converter.Converter) *serviceImpl {
	return &serviceImpl{
		dbStore:    dbStore,
		cacheStore: cacheStore,
		conv:       conv,
	}
}

type serviceImpl struct {
	dbStore    db.Store
	cacheStore cache.Store
	conv       converter.Converter
}

// RedirectTo returns the original url of the urlID if exists.
func (s *serviceImpl) RedirectTo(ctx context.Context, urlID string) (string, error) {
	shortURL, err := s.getShortURL(ctx, urlID)
	if err != nil {
		return "", err
	}
	if util.IsRecordExpired(shortURL) {
		return "", util.ErrURLExpired
	}
	return shortURL.URL, nil
}

func (s *serviceImpl) getShortURL(ctx context.Context, urlID string) (*record.ShortURL, error) {
	var err error
	var id int64
	var shortURL *record.ShortURL

	shortURL, err = s.cacheStore.Get(ctx, urlID)
	if err == nil {
		// cache hit, return the result
		return shortURL, nil
	}
	if err != cache.ErrKeyNotFound {
		// suppress error
		log.Errorf("getShortURL: cache store err: %v, with url_id: %v", err, urlID)
	}
	id, err = s.conv.ConvertToID(urlID)
	if err != nil {
		return nil, err
	}
	shortURL, err = s.dbStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return shortURL, nil
}
