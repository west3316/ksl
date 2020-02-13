package ksl

import (
	"sync"
)

type opContext struct {
	sync.Mutex
	Value interface{}
}

type opMgr struct {
	ops sync.Map
}

func (c *opMgr) lockValueOp(key string) {
	v, _ := c.ops.LoadOrStore(key, &opContext{})
	v.(*opContext).Lock()
}

func (c *opMgr) unlockValueOp(key string) {
	if v, exist := c.ops.Load(key); exist {
		v.(*opContext).Unlock()
	}
}

func (c *opMgr) updateValue(key string, fnUpdate func(value interface{}) interface{}) bool {
	v := c.loadValue(key)
	v = fnUpdate(v)
	if v == nil {
		return false
	}

	c.storeValue(key, v)
	return true
}

func (c *opMgr) storeValue(key string, value interface{}) {
	v, exist := c.ops.LoadOrStore(key, &opContext{Value: value})
	if exist {
		v.(*opContext).Value = value
		c.ops.Store(key, v)
	}
}

func (c *opMgr) loadValue(key string) interface{} {
	if v, exist := c.ops.Load(key); exist {
		return v.(*opContext).Value
	}
	return nil
}

func (c *opMgr) removeOp(key string) {
	c.ops.Delete(key)
}
