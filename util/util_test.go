package util

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	md "github.com/thegodmouse/url-shortener/db/mock"
	"github.com/thegodmouse/url-shortener/db/record"
)

func TestIsRecordDeleted(t *testing.T) {

	testCases := []struct {
		shortURL *record.ShortURL
		expBool  bool
	}{
		{
			shortURL: nil,
			expBool:  false,
		},
		{
			shortURL: &record.ShortURL{},
			expBool:  false,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-time.Minute),
				ExpireAt:  time.Now().Add(time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: false,
			},
			expBool: false,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-2 * time.Minute),
				ExpireAt:  time.Now().Add(-time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: false,
			},
			expBool: false,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-time.Minute),
				ExpireAt:  time.Now().Add(time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: true,
			},
			expBool: true,
		},
	}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.expBool, IsRecordDeleted(testCase.shortURL))
	}
}

func TestIsRecordExpired(t *testing.T) {

	testCases := []struct {
		shortURL *record.ShortURL
		expBool  bool
	}{
		{
			shortURL: nil,
			expBool:  false,
		},
		{
			shortURL: &record.ShortURL{},
			expBool:  true,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-time.Minute),
				ExpireAt:  time.Now().Add(time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: false,
			},
			expBool: false,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-2 * time.Minute),
				ExpireAt:  time.Now().Add(-time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: false,
			},
			expBool: true,
		},
		{
			shortURL: &record.ShortURL{
				ID:        int64(123),
				CreatedAt: time.Now().Add(-time.Minute),
				ExpireAt:  time.Now().Add(time.Minute),
				URL:       "http://localhost:5678",
				IsDeleted: true,
			},
			expBool: false,
		},
	}
	for _, testCase := range testCases {
		assert.Equal(t, testCase.expBool, IsRecordExpired(testCase.shortURL))
	}
}

type DeleteExpiredURLsTestSuite struct {
	suite.Suite

	ctrl *gomock.Controller

	dbStore *md.MockStore
}

func TestDeleteExpiredURLsSuite(t *testing.T) {
	suite.Run(t, new(DeleteExpiredURLsTestSuite))
}

func (s *DeleteExpiredURLsTestSuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
}

func (s *DeleteExpiredURLsTestSuite) SetupTest() {
	s.dbStore = md.NewMockStore(s.ctrl)
}

func (s *DeleteExpiredURLsTestSuite) TestDeleteExpiredURLs() {

	expCh := make(chan int64, 3)
	for i := 1; i <= 3; i++ {
		expCh <- int64(i)
	}
	close(expCh)

	ctx, cancel := context.WithCancel(context.Background())

	s.dbStore.
		EXPECT().
		GetExpiredIDs(gomock.Any()).
		Do(func(_ context.Context) { cancel() }).
		Return(expCh, nil)

	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(1)).
		Return(nil)
	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(2)).
		Return(nil)
	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(3)).
		Return(nil)

	<-DeleteExpiredURLs(ctx, s.dbStore, 500*time.Millisecond)
}

func (s *DeleteExpiredURLsTestSuite) TestDeleteExpiredURLs_withGetExpiredIdsError() {

	ctx, cancel := context.WithCancel(context.Background())

	s.dbStore.
		EXPECT().
		GetExpiredIDs(gomock.Any()).
		Do(func(_ context.Context) { cancel() }).
		Return(nil, errors.New("unknown query error"))

	<-DeleteExpiredURLs(ctx, s.dbStore, 500*time.Millisecond)
}

func (s *DeleteExpiredURLsTestSuite) TestDeleteExpiredURLs_withExpireError() {

	expCh := make(chan int64, 3)
	for i := 1; i <= 3; i++ {
		expCh <- int64(i)
	}
	close(expCh)

	ctx, cancel := context.WithCancel(context.Background())

	s.dbStore.
		EXPECT().
		GetExpiredIDs(gomock.Any()).
		Do(func(_ context.Context) { cancel() }).
		Return(expCh, nil)

	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(1)).
		Return(errors.New("unknown expire error"))
	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(2)).
		Return(nil)
	s.dbStore.
		EXPECT().
		Expire(gomock.Any(), int64(3)).
		Return(nil)

	<-DeleteExpiredURLs(ctx, s.dbStore, 500*time.Millisecond)
}
