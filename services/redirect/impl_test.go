package redirect

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/thegodmouse/url-shortener/cache"
	mc "github.com/thegodmouse/url-shortener/cache/mock"
	"github.com/thegodmouse/url-shortener/db"
	md "github.com/thegodmouse/url-shortener/db/mock"
	"github.com/thegodmouse/url-shortener/db/record"
)

func TestRedirectSuite(t *testing.T) {
	suite.Run(t, new(RedirectTestSuite))
}

type RedirectTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller

	mockCache *mc.MockStore
	mockDB    *md.MockStore
}

func (s *RedirectTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *RedirectTestSuite) SetupTest() {
	s.mockCache = mc.NewMockStore(s.ctrl)
	s.mockDB = md.NewMockStore(s.ctrl)
}

func (s *RedirectTestSuite) TestRedirectTo_withCacheHit() {
	srv := NewService(s.mockDB, s.mockCache)

	id := int64(12345)
	expURL := "http://localhost:5678"
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Minute),
		ExpireAt:  time.Now().Add(time.Minute),
		URL:       expURL,
	}

	s.mockCache.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(shortURL, nil)

	// SUT
	gotURL, gotErr := srv.RedirectTo(context.Background(), id)

	s.NoError(gotErr)
	s.Equal(expURL, gotURL)
}

func (s *RedirectTestSuite) TestRedirectTo_withCacheMissDatabaseFound() {
	srv := NewService(s.mockDB, s.mockCache)

	id := int64(54321)
	expURL := "http://localhost:5678"
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Minute),
		ExpireAt:  time.Now().Add(time.Minute),
		URL:       expURL,
	}

	s.mockCache.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.mockDB.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(shortURL, nil)

	// SUT
	gotURL, gotErr := srv.RedirectTo(context.Background(), id)

	s.NoError(gotErr)
	s.Equal(expURL, gotURL)
}

func (s *RedirectTestSuite) TestRedirectTo_withCacheError() {
	srv := NewService(s.mockDB, s.mockCache)

	id := int64(54321)
	expURL := "http://localhost:5678"
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Minute),
		ExpireAt:  time.Now().Add(time.Minute),
		URL:       expURL,
	}

	s.mockCache.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, errors.New("unknown cache error"))
	s.mockDB.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(shortURL, nil)

	// SUT
	gotURL, gotErr := srv.RedirectTo(context.Background(), id)

	s.NoError(gotErr)
	s.Equal(expURL, gotURL)
}

func (s *RedirectTestSuite) TestRedirectTo_withURLNotFound() {
	srv := NewService(s.mockDB, s.mockCache)

	id := int64(54321)

	s.mockCache.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.mockDB.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, db.ErrNoRows)

	// SUT
	gotURL, gotErr := srv.RedirectTo(context.Background(), id)

	s.Error(gotErr)
	s.Empty(gotURL)
}
