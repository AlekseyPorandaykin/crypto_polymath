package dispatcher

import "time"

type Event[T interface{}] struct {
	Name string
	Body T
}

type Listener[T interface{}] interface {
	Handle(e Event[T])
}

type Dispatcher[T interface{}] struct {
	listeners   []Listener[T]
	preHandler  func(e Event[T])
	postHandler func(e Event[T], duration time.Duration)
	closeCh     chan struct{}
	events      chan Event[T]
}

func New[T interface{}]() *Dispatcher[T] {
	return &Dispatcher[T]{
		listeners: make([]Listener[T], 0),
		closeCh:   make(chan struct{}, 1),
		events:    make(chan Event[T], 10_000),
	}
}

func (c *Dispatcher[T]) SetPreHandler(handler func(e Event[T])) {
	c.preHandler = handler
}

func (c *Dispatcher[T]) SetPostHandler(handler func(e Event[T], duration time.Duration)) {
	c.postHandler = handler
}

func (c *Dispatcher[T]) Dispatch(event Event[T]) {
	c.events <- event
}

func (c *Dispatcher[T]) Subscribe(consumer Listener[T]) {
	c.listeners = append(c.listeners, consumer)
}

func (c *Dispatcher[T]) Listen() {
	for {
		select {
		case <-c.closeCh:
			return
		case e := <-c.events:
			for _, consumer := range c.listeners {
				now := time.Now()
				if c.preHandler != nil {
					c.preHandler(e)
				}
				consumer.Handle(e)
				if c.postHandler != nil {
					c.postHandler(e, time.Now().Sub(now))
				}
			}
		}
	}
}

func (c *Dispatcher[T]) Close() {
	c.closeCh <- struct{}{}
}
