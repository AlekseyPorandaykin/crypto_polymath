package cache

import "sync"

type TwoQueue[F comparable, S comparable, T interface{}] struct {
	data map[F]map[S]T
	mu   sync.Mutex
}

func NewTwoQueueCache[F comparable, S comparable, T interface{}]() *TwoQueue[F, S, T] {
	return &TwoQueue[F, S, T]{
		data: make(map[F]map[S]T),
	}
}

func (c *TwoQueue[F, S, T]) Set(firstKey F, secondKey S, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data[firstKey] == nil {
		c.data[firstKey] = make(map[S]T)
	}
	c.data[firstKey][secondKey] = val
}

func (c *TwoQueue[F, S, T]) Get(firstKey F, secondKey S) *T {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data[firstKey] == nil {
		return nil
	}
	data, has := c.data[firstKey][secondKey]
	if has {
		return &data
	}
	return nil
}

func (c *TwoQueue[F, S, T]) ValuesByFirstKey(firstKey F) []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data[firstKey] == nil || len(c.data[firstKey]) == 0 {
		return nil
	}
	result := make([]T, 0, len(c.data[firstKey]))
	for _, item := range c.data[firstKey] {
		result = append(result, item)
	}
	return result
}

func (c *TwoQueue[F, S, T]) ValuesBySecondKey(secondKey S) []T {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]T, 0, 1_000)
	for _, values := range c.data {
		if _, has := values[secondKey]; has {
			result = append(result, values[secondKey])
		}
	}
	return result
}

func (c *TwoQueue[F, S, T]) Values() []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]T, 0, 1_000)
	for _, values := range c.data {
		for _, value := range values {
			result = append(result, value)
		}
	}
	return result
}

func (c *TwoQueue[F, S, T]) ValuesByCondition(condition func(item T) bool) []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]T, 0, 1_000)
	for _, values := range c.data {
		for _, value := range values {
			if condition(value) {
				result = append(result, value)
			}
		}
	}
	return result
}
func (c *TwoQueue[F, S, T]) ListByCondition(firstKey F, secondKey S, condition func(item T) bool) []T {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]T, 0, 1_000)
	for _, values := range c.data {
		for _, value := range values {
			if condition(value) {
				result = append(result, value)
			}
		}
	}
	return result
}

func (c *TwoQueue[F, S, T]) DeleteByCondition(condition func(item T) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for firstKey, values := range c.data {
		for secondKey, value := range values {
			if condition(value) {
				delete(c.data[firstKey], secondKey)
			}
		}
	}
}
