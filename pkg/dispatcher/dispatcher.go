package dispatcher

import "time"

type Event[T interface{}] struct {
	Name string
	Body T
}

type Listener[T interface{}] interface {
	Handle(e Event[T])
}

type Dispatcher[T interface{}] interface {
	SetPreHandler(handler func(e Event[T]))
	SetPostHandler(handler func(e Event[T], duration time.Duration))
	SetPreDispatcher(handler func(e Event[T]))
	SetPostDispatcher(handler func(e Event[T]))
	Dispatch(e Event[T])
	Subscribe(consumer Listener[T])
	Listen()
	Close()
}
