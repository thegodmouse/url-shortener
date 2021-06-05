package util

import (
	"errors"
	"time"

	"github.com/thegodmouse/url-shortener/db/record"
)

var (
	ErrURLFormat  = errors.New("urlID is in wrong format")
	ErrURLExpired = errors.New("short url is expired")
)

func IsRecordExpired(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return true
	}
	return shortURL.ExpireAt.Before(time.Now())
}
