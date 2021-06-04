package util

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrURLFormat = errors.New("urlID is in wrong format")
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
