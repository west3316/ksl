package ksl

import (
	"sync"
	"sync/atomic"
)

// BidiMap goroutine safe bidimap with len
type BidiMap struct {
	m1 sync.Map
	m2 sync.Map
	l  int32
}

// GetOrPut returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (c *BidiMap) GetOrPut(k, v interface{}) (val interface{}, loaded bool) {
	val, loaded = c.m1.LoadOrStore(k, v)
	if !loaded {
		c.m2.LoadOrStore(v, k)
	}
	return
}

// Put put key and value in bidimap
func (c *BidiMap) Put(key interface{}, v interface{}) {
	c.m1.Store(key, v)
	c.m2.Store(v, key)
	atomic.AddInt32(&c.l, 1)
}

// GetByKey get value by key
func (c *BidiMap) GetByKey(key interface{}) (interface{}, bool) {
	return c.m1.Load(key)
}

// GetByValue get key by value
func (c *BidiMap) GetByValue(value interface{}) (interface{}, bool) {
	return c.m2.Load(value)
}

// ExistKey return true when key exist, otherwise false
func (c *BidiMap) ExistKey(key interface{}) bool {
	_, exist := c.m1.Load(key)
	return exist
}

// ExistValue return true when value exist, otherwise false
func (c *BidiMap) ExistValue(value interface{}) bool {
	_, exist := c.m2.Load(value)
	return exist
}

// RemoveByValue remove bidimap element by value
func (c *BidiMap) RemoveByValue(value interface{}) interface{} {
	key, exist := c.GetByValue(value)
	if !exist {
		return nil
	}

	c.m1.Delete(key)
	c.m2.Delete(value)
	atomic.AddInt32(&c.l, -1)
	return key
}

// RemoveByKey remove bidimap element by key
func (c *BidiMap) RemoveByKey(key interface{}) interface{} {
	value, exist := c.GetByKey(key)
	if !exist {
		return nil
	}

	c.m1.Delete(key)
	c.m2.Delete(value)
	atomic.AddInt32(&c.l, -1)
	return value
}

// Len return map len
func (c *BidiMap) Len() int {
	return int(atomic.LoadInt32(&c.l))
}

// Foreach loop bidimap
func (c *BidiMap) Foreach(fn func(key interface{}, v interface{}) bool) {
	c.m1.Range(func(key interface{}, val interface{}) bool {
		return fn(key, val)
	})
}
