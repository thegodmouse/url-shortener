package converter

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
)

var (
	// ErrURLFormat is returned when the given url format cannot be converted.
	ErrURLFormat = errors.New("url id is in wrong format")
)

// Converter defines the interface for the conversion between id and short url id.
// These two functions should be inverses of each other.
type Converter interface {
	// ConvertToID converts
	ConvertToID(shortURLID string) (int64, error)
	ConvertToShortURL(id int64) (string, error)
}

// NewConverter returns a default converter which implements Converter
func NewConverter() *converterImpl {
	return &converterImpl{}
}

type converterImpl struct{}

// ConvertToShortURL converts an id to the unique short url id.
func (c *converterImpl) ConvertToShortURL(id int64) (string, error) {
	return fmt.Sprintf("%d", id), nil
}

// ConvertToID converts a short url id to the unique id.
func (c *converterImpl) ConvertToID(shortURLID string) (int64, error) {
	id, err := strconv.ParseInt(shortURLID, 10, 64)
	if err != nil {
		log.Errorf("converterImpl.ConvertToID: convert err: %v, with short url id: %v", err, shortURLID)
		return 0, ErrURLFormat
	}
	return id, err
}
