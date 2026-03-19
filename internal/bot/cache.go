package bot

import (
	"container/list"
	"sync"
)

type boundedCache[T any] struct {
	mu    sync.Mutex
	max   int
	ll    *list.List
	items map[string]*list.Element
}

type cacheEntry[T any] struct {
	key   string
	value T
}

func newBoundedCache[T any](max int) *boundedCache[T] {
	if max <= 0 {
		max = 1
	}
	return &boundedCache[T]{
		max:   max,
		ll:    list.New(),
		items: make(map[string]*list.Element, max),
	}
}

func (c *boundedCache[T]) Get(key string) (T, bool) {
	var zero T
	if c == nil {
		return zero, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return zero, false
	}
	c.ll.MoveToFront(elem)
	return elem.Value.(cacheEntry[T]).value, true
}

func (c *boundedCache[T]) Set(key string, value T) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		elem.Value = cacheEntry[T]{key: key, value: value}
		c.ll.MoveToFront(elem)
		return
	}
	elem := c.ll.PushFront(cacheEntry[T]{key: key, value: value})
	c.items[key] = elem
	if c.ll.Len() <= c.max {
		return
	}
	tail := c.ll.Back()
	if tail == nil {
		return
	}
	c.ll.Remove(tail)
	delete(c.items, tail.Value.(cacheEntry[T]).key)
}

func (c *boundedCache[T]) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}
