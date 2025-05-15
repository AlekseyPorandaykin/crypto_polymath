package dispatcher

import (
	"time"
)

type EmptyDispatcher[T any] struct{}

func NewEmptyDispatcher[T any]() Dispatcher[T] {
	return &EmptyDispatcher[T]{}
}

func (d *EmptyDispatcher[T]) SetPreHandler(handler func(e Event[T])) {
	return
}

func (d *EmptyDispatcher[T]) SetPostHandler(handler func(e Event[T], duration time.Duration)) {
	return
}

func (d *EmptyDispatcher[T]) SetPreDispatcher(handler func(e Event[T])) {
	return
}

func (d *EmptyDispatcher[T]) SetPostDispatcher(handler func(e Event[T])) {
	return
}

func (d *EmptyDispatcher[T]) Dispatch(e Event[T]) {
	return
}

func (d *EmptyDispatcher[T]) Subscribe(consumer Listener[T]) {
	return
}

func (d *EmptyDispatcher[T]) Listen() {
	return
}

func (d *EmptyDispatcher[T]) Close() {
	return
}
