package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
