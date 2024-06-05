package dispatcher

type Event[T interface{}] struct {
	Name string
	Body T
}

type Listener[T interface{}] interface {
	Handle(e Event[T])
}

type Dispatcher[T interface{}] struct {
	listeners []Listener[T]
	closeCh   chan struct{}
	events    chan Event[T]
}

func New[T interface{}]() *Dispatcher[T] {
	return &Dispatcher[T]{
		listeners: make([]Listener[T], 0),
		closeCh:   make(chan struct{}, 1),
		events:    make(chan Event[T], 100),
	}
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
				consumer.Handle(e)
			}
		}
	}
}

func (c *Dispatcher[T]) Close() {
	c.closeCh <- struct{}{}
}
