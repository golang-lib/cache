// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"time"
)

// Value that go into LRUCache need to satisfy this interface.
type Value interface {
	Size() int
}

type Item struct {
	Key   string
	Value interface{}
	Size  int
}

type entry struct {
	key            string
	value          interface{}
	size           int
	finalizer      interface{}
	timeExpiration time.Time
	timeAccessed   time.Time
}

func (i *entry) Expired() bool {
	return i.timeExpiration.Before(time.Now())
}

func (i *entry) Finalize() {
	if i.finalizer == nil {
		return
	}
	if f, ok := i.finalizer.(func()); ok {
		i.finalizer = nil
		f()
		return
	}
	if f, ok := i.finalizer.(func(x interface{})); ok {
		i.finalizer = nil
		f(i.value)
		return
	}
	if f, ok := i.finalizer.(func(x Value)); ok {
		i.finalizer = nil
		f(i.value.(Value))
		return
	}
	panic("not reachable")
}
