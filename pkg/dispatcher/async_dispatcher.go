package dispatcher

import "time"

type AsyncDispatcher[T interface{}] struct {
	listeners      []Listener[T]
	preHandler     func(e Event[T])
	postHandler    func(e Event[T], duration time.Duration)
	preDispatcher  func(e Event[T])
	postDispatcher func(e Event[T])
	closeCh        chan struct{}
	events         chan Event[T]
}

func NewAsyncDispatcher[T interface{}]() Dispatcher[T] {
	return &AsyncDispatcher[T]{
		listeners: make([]Listener[T], 0),
		closeCh:   make(chan struct{}, 1),
		events:    make(chan Event[T], 100),
	}
}

func (c *AsyncDispatcher[T]) SetPreHandler(handler func(e Event[T])) {
	c.preHandler = handler
}

func (c *AsyncDispatcher[T]) SetPostHandler(handler func(e Event[T], duration time.Duration)) {
	c.postHandler = handler
}

func (c *AsyncDispatcher[T]) SetPreDispatcher(handler func(e Event[T])) {
	c.preDispatcher = handler
}
func (c *AsyncDispatcher[T]) SetPostDispatcher(handler func(e Event[T])) {
	c.postDispatcher = handler
}

func (c *AsyncDispatcher[T]) Dispatch(e Event[T]) {
	if c.preDispatcher != nil {
		c.preDispatcher(e)
	}
	c.events <- e
	if c.postDispatcher != nil {
		c.postDispatcher(e)
	}
}

func (c *AsyncDispatcher[T]) Subscribe(consumer Listener[T]) {
	c.listeners = append(c.listeners, consumer)
}

func (c *AsyncDispatcher[T]) Listen() {
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

func (c *AsyncDispatcher[T]) Close() {
	c.closeCh <- struct{}{}
}
