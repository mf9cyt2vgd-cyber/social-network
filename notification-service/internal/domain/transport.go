package domain

import "context"

type EventConsumer interface {
	Consume(ctx context.Context, c chan []byte, cErr chan error)
}
