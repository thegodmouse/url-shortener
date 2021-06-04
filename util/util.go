package util

import (
	"errors"
	"fmt"
	"github.com/thegodmouse/url-shortener/db/record"
	"strconv"
	"time"
)

var (
	ErrURLFormat  = errors.New("urlID is in wrong format")
	ErrURLExpired = errors.New("short url is expired")
)

func ConvertToShortURL(id int64) (string, error) {
	return fmt.Sprintf("%06d", id), nil
}

func ConvertToID(shortURL string) (int64, error) {
	id, err := strconv.ParseInt(shortURL, 10, 64)
	if err != nil {
		return 0, ErrURLFormat
	}
	return id, err
}

func IsRecordExpired(shortURL *record.ShortURL) bool {
	if shortURL == nil {
		return true
	}
	return shortURL.ExpireAt.Before(time.Now())
}
