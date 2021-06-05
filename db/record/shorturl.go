package record

import "time"

type ShortURL struct {
	ID        int64
	CreatedAt time.Time
	ExpireAt  time.Time
	URL       string
	IsDeleted bool
}
