package redirect

import (
	"context"
	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
	"github.com/thegodmouse/url-shortener/util"
)

type Service interface {
	RedirectTo(ctx context.Context, urlID string) (string, error)
}

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

func (s *serviceImpl) RedirectTo(ctx context.Context, urlID string) (string, error) {
	var err error
	var id int64
	var shortURL *record.ShortURL

	shortURL, err = s.cacheStore.Get(ctx, urlID)
	if err == nil {
		return shortURL.URL, nil
	}
	id, err = util.ConvertToID(urlID)
	if err != nil {
		return "", err
	}
	shortURL, err = s.dbStore.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return shortURL.URL, nil
}
