package util

import (
	"fmt"
	"strconv"
)

func ConvertToShortURL(id int64) (string, error) {
	return fmt.Sprintf("%06d", id), nil
}

func ConvertToID(shortURL string) (int64, error) {
	return strconv.ParseInt(shortURL, 10, 64)
}
