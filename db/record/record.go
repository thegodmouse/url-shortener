package record

import (
	"encoding/json"
	"time"
)

type ShortURL struct {
	ID        int64
	CreatedAt time.Time
	ExpireAt  time.Time
	URL       string
	IsDeleted bool
}

func (r *ShortURL) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *ShortURL) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}

type RecycleURL struct {
	ID        int64
	CreatedAt time.Time
}

func (r *RecycleURL) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *RecycleURL) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, r)
}
