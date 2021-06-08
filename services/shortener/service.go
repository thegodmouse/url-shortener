package shortener

import (
	"context"
	"time"
)

// Service defines the interface for shortening and deleting urls.
type Service interface {
	// Shorten shortens an url with an unique id, and create a record in the database.
	Shorten(ctx context.Context, url string, expireAt time.Time) (int64, error)
	// Delete deletes an url with id.
	Delete(ctx context.Context, id int64) error
}
