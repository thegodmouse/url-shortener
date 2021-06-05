package redirect

import (
	"context"
)

type Service interface {
	RedirectTo(ctx context.Context, id int64) (string, error)
}
