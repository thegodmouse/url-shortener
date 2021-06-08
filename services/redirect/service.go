package redirect

import (
	"context"
)

// Service defines the interface for redirect url with id.
type Service interface {
	// RedirectTo returns the original url with given id.
	RedirectTo(ctx context.Context, id int64) (string, error)
}
