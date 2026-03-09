package domain

import "context"

type EventConsumer interface {
	Consume(ctx context.Context) chan *Post
}
