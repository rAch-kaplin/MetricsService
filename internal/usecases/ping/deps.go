package ping

import (
	"context"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Closer interface {
	Close() error
}
