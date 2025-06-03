package queue

import "context"

type Publisher[T any] interface {
	Publish(message ...T) error
}

type Consumer[T any] interface {
	Consume(ctx context.Context, handler func(message T) error) error
}

type Receiver[T any] interface {
	Listen()
	Close()
	Receive() (*T, error)
}
