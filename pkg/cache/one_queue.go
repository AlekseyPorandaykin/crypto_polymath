package cache

import (
	"context"
	"sync"
)

type OneQueue[K comparable, T interface{}] struct {
	data map[K]T
	mu   sync.Mutex
}

func NewCache[K comparable, T interface{}]() *OneQueue[K, T] {
	return &OneQueue[K, T]{data: make(map[K]T)}
}

func (c *OneQueue[K, T]) Set(ctx context.Context, key K, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}

func (c *OneQueue[K, T]) Get(ctx context.Context, key K) *T {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, has := c.data[key]
	if has {
		return &val
	}
	return nil
}
