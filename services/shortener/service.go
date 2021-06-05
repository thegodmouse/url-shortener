package shortener

import (
	"context"
	"fmt"
	"time"

	"github.com/thegodmouse/url-shortener/cache"
	"github.com/thegodmouse/url-shortener/converter"
	"github.com/thegodmouse/url-shortener/db"
	"github.com/thegodmouse/url-shortener/db/record"
)

type Service interface {
	Shorten(ctx context.Context, url string, expireAt time.Time) (string, error)
	Delete(ctx context.Context, urlID string) error
}

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

func (s *serviceImpl) Shorten(ctx context.Context, url string, expireAt time.Time) (string, error) {
	var err error
	var shortURL *record.ShortURL
	var urlID string

	shortURL, err = s.dbStore.Create(ctx, url, expireAt)
	if err != nil {
		return "", err
	}
	urlID, err = s.conv.ConvertToShortURL(shortURL.ID)
	if err != nil {
		return "", err
	}

	err = s.cacheStore.Set(ctx, urlID, shortURL)
	if err != nil {
		fmt.Printf("QQ set err:%v\n", err)
	}
	return "", nil
}

func (s *serviceImpl) Delete(ctx context.Context, urlID string) error {
	id, err := s.conv.ConvertToID(urlID)
	if err != nil {
		return err
	}
	err = s.dbStore.Delete(ctx, id)
	if err != nil {
		return err
	}
	err = s.cacheStore.Evict(ctx, urlID)
	if err != nil {
		fmt.Printf("QQ evict err:%v\n", err)
	}
	return nil
}
