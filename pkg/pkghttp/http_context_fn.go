package pkghttp

import (
	"context"
)

type ContextFn func(ctx context.Context, r Request) context.Context
