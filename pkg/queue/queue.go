package queue

import "context"

type KeyEvent interface {
	KeyEvent() string
}

type Publisher[T any] interface {
	Publish(message ...T) error
}

type Consumer[T any] interface {
	Consume(ctx context.Context, handler func(message T) error) error
}

type Receiver[T any] interface {
	Listen()
	Close()
	Receive(context.Context) ([]*T, error)
}
