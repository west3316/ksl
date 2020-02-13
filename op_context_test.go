package ksl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValue struct {
	Data string
}

func TestOpContext(t *testing.T) {
	mgr := &opMgr{}
	mgr.storeValue("a", &testValue{Data: "123"})
	mgr.storeValue("a", &testValue{Data: "456"})
	d2 := mgr.loadValue("a")
	assert.Equal(t, "456", d2.(*testValue).Data)
}
