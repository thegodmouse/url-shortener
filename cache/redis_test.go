package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/suite"
	"github.com/thegodmouse/url-shortener/db/record"
)

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

type RedisTestSuite struct {
	suite.Suite

	cache *redis.Client
	mock  redismock.ClientMock
}

func (s *RedisTestSuite) SetupTest() {
	s.cache, s.mock = redismock.NewClientMock()
}

func (s *RedisTestSuite) TearDownTest() {
	s.mock.ExpectationsWereMet()
}

func (s *RedisTestSuite) TestNewRedisStore() {
	redisStore := newRedisStore(s.cache)

	s.Equal(defaultExpiration, redisStore.expiration)
}

func (s *RedisTestSuite) TestGetHit() {
	redisStore := newRedisStore(s.cache)

	id := int64(12345)
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Hour).Round(time.Second),
		ExpireAt:  time.Now().Add(time.Hour).Round(time.Second),
		URL:       "http://localhost:6789",
		IsDeleted: false,
	}
	data, _ := shortURL.MarshalBinary()

	s.mock.
		ExpectGet(redisStore.makeKey(id)).
		SetVal(string(data))

	// SUT
	gotRecord, gotErr := redisStore.Get(context.Background(), id)

	s.NoError(gotErr)
	s.Equal(shortURL.ID, gotRecord.ID)
	s.Equal(shortURL.URL, gotRecord.URL)
	s.Equal(shortURL.CreatedAt, gotRecord.CreatedAt)
	s.Equal(shortURL.ExpireAt, gotRecord.ExpireAt)
	s.Equal(shortURL.IsDeleted, gotRecord.IsDeleted)
}

func (s *RedisTestSuite) TestGetMiss() {
	redisStore := newRedisStore(s.cache)

	id := int64(12345)

	s.mock.
		ExpectGet(redisStore.makeKey(id)).
		RedisNil()

	// SUT
	gotRecord, gotErr := redisStore.Get(context.Background(), id)

	s.Error(gotErr)
	s.Equal(redis.Nil, gotErr)
	s.Nil(gotRecord)
}

func (s *RedisTestSuite) TestGetError() {
	redisStore := newRedisStore(s.cache)

	id := int64(12345)

	s.mock.
		ExpectGet(redisStore.makeKey(id)).
		SetErr(errors.New("unknown get error"))

	// SUT
	gotRecord, gotErr := redisStore.Get(context.Background(), id)

	s.Error(gotErr)
	s.NotEqual(redis.Nil, gotErr)
	s.Nil(gotRecord)
}

func (s *RedisTestSuite) TestSet() {
	redisStore := newRedisStore(s.cache)

	id := int64(12345)
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Hour).Round(time.Second),
		ExpireAt:  time.Now().Add(time.Hour).Round(time.Second),
		URL:       "http://localhost:6789",
		IsDeleted: false,
	}

	s.mock.
		ExpectSet(redisStore.makeKey(id), shortURL, redisStore.expiration).
		SetVal("OK")

	// SUT
	gotErr := redisStore.Set(context.Background(), id, shortURL)

	s.NoError(gotErr)
}

func (s *RedisTestSuite) TestSetError() {
	redisStore := newRedisStore(s.cache)

	id := int64(12345)
	shortURL := &record.ShortURL{
		ID:        id,
		CreatedAt: time.Now().Add(-time.Hour).Round(time.Second),
		ExpireAt:  time.Now().Add(time.Hour).Round(time.Second),
		URL:       "http://localhost:6789",
		IsDeleted: false,
	}

	s.mock.
		ExpectSet(redisStore.makeKey(id), shortURL, redisStore.expiration).
		SetErr(errors.New("unknown set error"))

	// SUT
	gotErr := redisStore.Set(context.Background(), id, shortURL)

	s.Error(gotErr)
}

func (s *RedisTestSuite) TestMakeKey() {
	redisStore := newRedisStore(s.cache)

	testCases := []struct {
		id     int64
		expKey string
	}{
		{
			id:     int64(0),
			expKey: "id#0",
		},
		{
			id:     int64(123),
			expKey: "id#123",
		},
		{
			id:     int64(-456),
			expKey: "id#-456",
		},
	}

	for _, testCase := range testCases {
		s.Equal(testCase.expKey, redisStore.makeKey(testCase.id))
	}
}
