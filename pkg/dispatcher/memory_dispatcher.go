package dispatcher

import (
	"sync"
	"time"
)

type MemoryDispatcher[T interface{}] struct {
	listeners          []Listener[T]
	preHandler         func(e Event[T])
	postHandler        func(e Event[T], duration time.Duration)
	preDispatcher      func(e Event[T])
	postDispatcher     func(e Event[T])
	closeCh            chan struct{}
	countAsyncHandlers int

	eventStorage []Event[T]
	muStorage    sync.Mutex
}

func NewMemoryDispatcher[T interface{}]() Dispatcher[T] {
	return &MemoryDispatcher[T]{
		listeners:    make([]Listener[T], 0),
		closeCh:      make(chan struct{}, 1),
		eventStorage: make([]Event[T], 0, 100),
	}
}

func (c *MemoryDispatcher[T]) SetPreHandler(handler func(e Event[T])) {
	c.preHandler = handler
}

func (c *MemoryDispatcher[T]) SetPostHandler(handler func(e Event[T], duration time.Duration)) {
	c.postHandler = handler
}

func (c *MemoryDispatcher[T]) SetPreDispatcher(handler func(e Event[T])) {
	c.preDispatcher = handler
}
func (c *MemoryDispatcher[T]) SetPostDispatcher(handler func(e Event[T])) {
	c.postDispatcher = handler
}

func (c *MemoryDispatcher[T]) Dispatch(e Event[T]) {
	if c.preDispatcher != nil {
		c.preDispatcher(e)
	}
	c.muStorage.Lock()
	c.eventStorage = append(c.eventStorage, e)
	c.muStorage.Unlock()
	if c.postDispatcher != nil {
		c.postDispatcher(e)
	}
}

func (c *MemoryDispatcher[T]) Subscribe(consumer Listener[T]) {
	c.listeners = append(c.listeners, consumer)
}

func (c *MemoryDispatcher[T]) Listen() {
	for {
		select {
		case <-c.closeCh:
			return
		default:
			e := c.firstEvent()
			if e == nil {
				continue
			}
			for _, consumer := range c.listeners {
				now := time.Now()
				if c.preHandler != nil {
					c.preHandler(*e)
				}
				consumer.Handle(*e)
				if c.postHandler != nil {
					c.postHandler(*e, time.Now().Sub(now))
				}
			}
		}
	}
}
func (c *MemoryDispatcher[T]) firstEvent() *Event[T] {
	c.muStorage.Lock()
	defer c.muStorage.Unlock()
	if len(c.eventStorage) == 0 {
		return nil
	}
	e := c.eventStorage[0]
	if len(c.eventStorage) == 1 {
		c.eventStorage = make([]Event[T], 0, 100)
		return &e
	}
	c.eventStorage = c.eventStorage[1:]
	return &e

}

func (c *MemoryDispatcher[T]) Close() {
	c.closeCh <- struct{}{}
}
