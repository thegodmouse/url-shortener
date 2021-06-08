package record

import (
	"encoding/json"
	"time"
)

// ShortURL is a record for storing the information of a short url.
type ShortURL struct {
	ID        int64
	CreatedAt time.Time
	ExpireAt  time.Time
	URL       string
	IsDeleted bool
}

// MarshalBinary marshals the record to binary data in json format.
func (r *ShortURL) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

// UnmarshalBinary restores the record from the binary data in json format.
func (r *ShortURL) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}
