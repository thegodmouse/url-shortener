package shortener

import (
	"context"
	"errors"
	"fmt"
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

func TestShortenerSuite(t *testing.T) {
	suite.Run(t, new(ShortenerTestSuite))
}

type ShortenerTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller

	dbStore    *md.MockStore
	cacheStore *mc.MockStore
}

func (s *ShortenerTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *ShortenerTestSuite) SetupTest() {
	s.dbStore = md.NewMockStore(s.ctrl)
	s.cacheStore = mc.NewMockStore(s.ctrl)
}

func (s *ShortenerTestSuite) TestShorten() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(123)
	url := "http://localhost:5678"
	createdAt := time.Now().Add(-time.Minute).Round(time.Second)
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		URL:       url,
	}

	s.dbStore.
		EXPECT().
		Create(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return(shortURL, nil)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: shortURL}).
		Return(nil)

	// SUT
	gotID, gotErr := srv.Shorten(context.Background(), url, expireAt)

	s.NoError(gotErr)
	s.Equal(id, gotID)
}

func (s *ShortenerTestSuite) TestShorten_withDatabaseError() {
	srv := NewService(s.dbStore, s.cacheStore)

	url := "http://localhost:5678"
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	s.dbStore.
		EXPECT().
		Create(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return(nil, errors.New("db error"))

	// SUT
	gotID, gotErr := srv.Shorten(context.Background(), url, expireAt)

	s.Error(gotErr)
	s.Equal(int64(0), gotID)
}

func (s *ShortenerTestSuite) TestShorten_withCacheError() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(123)
	url := "http://localhost:5678"
	createdAt := time.Now().Add(-time.Minute).Round(time.Second)
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		URL:       url,
	}

	s.dbStore.
		EXPECT().
		Create(gomock.Any(), gomock.Eq(url), gomock.Eq(expireAt)).
		Return(shortURL, nil)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: shortURL}).
		Return(errors.New("cache error"))

	// SUT
	gotID, gotErr := srv.Shorten(context.Background(), url, expireAt)

	s.NoError(gotErr)
	s.Equal(id, gotID)
}

func (s *ShortenerTestSuite) TestDelete() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(nil)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: &record.ShortURL{ID: id, IsDeleted: true}}).
		Return(nil)

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withDatabaseError() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(errors.New("db error"))

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withCacheGetError() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, errors.New("unknown cache error"))
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(nil)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: &record.ShortURL{ID: id, IsDeleted: true}}).
		Return(nil)

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withCacheSetError() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(nil)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: &record.ShortURL{ID: id, IsDeleted: true}}).
		Return(errors.New("unknown cache err"))

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withRecordDeleted() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	url := "http://localhost:5678"
	createdAt := time.Now().Add(-time.Minute).Round(time.Second)
	expireAt := time.Now().Add(time.Minute).Round(time.Second)

	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		URL:       url,
		IsDeleted: true,
	}

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(shortURL, nil)

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withRecordNotExist_andCacheHit() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	shortURL := &record.ShortURL{
		ID:         id,
		IsNotExist: true,
	}

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(shortURL, nil)

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.NoError(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withRecordNotExist_andCacheMiss() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(db.ErrNoRows)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: &record.ShortURL{ID: id, IsNotExist: true}}).
		Return(nil)

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.Error(gotErr)
}

func (s *ShortenerTestSuite) TestDelete_withRecordNotExist_andCacheError() {
	srv := NewService(s.dbStore, s.cacheStore)

	id := int64(12345)

	s.cacheStore.
		EXPECT().
		Get(gomock.Any(), gomock.Eq(id)).
		Return(nil, cache.ErrKeyNotFound)
	s.dbStore.
		EXPECT().
		Delete(gomock.Any(), gomock.Eq(id)).
		Return(db.ErrNoRows)
	s.cacheStore.
		EXPECT().
		Set(gomock.Any(), gomock.Eq(id), &recordMatcher{shortURL: &record.ShortURL{ID: id, IsNotExist: true}}).
		Return(errors.New("unknown cache error"))

	// SUT
	gotErr := srv.Delete(context.Background(), id)

	s.Error(gotErr)
}

type recordMatcher struct {
	shortURL *record.ShortURL
}

func (m recordMatcher) Matches(x interface{}) bool {
	shortURL, ok := x.(*record.ShortURL)
	if !ok {
		return false
	}
	return m.shortURL.ID == shortURL.ID &&
		m.shortURL.URL == shortURL.URL &&
		m.shortURL.ExpireAt.Equal(shortURL.ExpireAt) &&
		m.shortURL.CreatedAt.Equal(shortURL.CreatedAt)
}

func (m recordMatcher) String() string {
	return fmt.Sprintf("has record: %+v", m.shortURL)
}
