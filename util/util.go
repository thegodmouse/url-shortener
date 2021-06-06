package util

import (
	"errors"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	ErrURLNotFound = errors.New("short url not found")
)

func IsRecordExpired(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.ExpireAt.Before(time.Now())
}

func IsRecordDeleted(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return false
	}
	return shortURL.IsDeleted
}
