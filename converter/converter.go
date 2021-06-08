package converter

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrURLFormat is returned when the given url format cannot be converted.
	ErrURLFormat = errors.New("url id is in wrong format")
)

// Converter defines the interface for the conversion between id and url id.
// These two functions should be inverses of each other.
type Converter interface {
	// ConvertToID converts
	ConvertToID(urlID string) (int64, error)
	ConvertToURLID(id int64) (string, error)
}

// NewConverter returns a default converter which implements Converter
func NewConverter() *converterImpl {
	return &converterImpl{}
}

type converterImpl struct{}

// ConvertToURLID converts an id to the unique short url id.
func (c *converterImpl) ConvertToURLID(id int64) (string, error) {
	return fmt.Sprintf("%d", id), nil
}

// ConvertToID converts an url id to the unique id.
func (c *converterImpl) ConvertToID(urlID string) (int64, error) {
	id, err := strconv.ParseInt(urlID, 10, 64)
	if err != nil {
		log.Errorf("converterImpl.ConvertToID: convert err: %v, with short url id: %v", err, urlID)
		return 0, ErrURLFormat
	}
	return id, err
}
