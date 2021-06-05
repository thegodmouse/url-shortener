package converter

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrURLFormat = errors.New("urlID is in wrong format")
)

type Converter interface {
	ConvertToID(shortURL string) (int64, error)
	ConvertToShortURL(id int64) (string, error)
}

func NewConverter() *converterImpl {
	return &converterImpl{}
}

type converterImpl struct{}

func (c *converterImpl) ConvertToShortURL(id int64) (string, error) {
	return fmt.Sprintf("%d", id), nil
}

func (c *converterImpl) ConvertToID(shortURL string) (int64, error) {
	id, err := strconv.ParseInt(shortURL, 10, 64)
	if err != nil {
		return 0, ErrURLFormat
	}
	return id, err
}
