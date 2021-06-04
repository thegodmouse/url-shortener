package redirect

import (
	"context"
)

type Service interface {
	RedirectTo(ctx context.Context, urlID string) (string, error)
}
