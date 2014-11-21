// Copyright 2013 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"runtime"
	"sync"
	"time"
)

type Cache struct {
	*cache
	x **cache // just for runtime.SetFinalizer
}

type cache struct {
	sync.Mutex
	defaultExpiration time.Duration
	table             map[string][]*entry
	janitor           *janitor
}

func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	if defaultExpiration <= 0 {
		defaultExpiration = time.Second * 30
	}
	if cleanupInterval <= 0 {
		cleanupInterval = time.Second * 1
	}

	c := &cache{
		defaultExpiration: defaultExpiration,
		table:             map[string][]*entry{},
	}
	x := &c

	startJanitor(c, cleanupInterval)
	runtime.SetFinalizer(x, stopJanitor)

	return &Cache{c, x}
}

func (c *cache) Fetch(k string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	table, found := c.table[k]
	if !found {
		return nil, false
	}

	var last *entry
	last, table[len(table)-1] = table[len(table)-1], nil
	table = table[:len(table)-1]

	if len(table) == 0 {
		delete(c.table, k)
	} else {
		c.table[k] = table
	}

	return last.value, true
}

func (c *cache) Store(k string, x interface{}, Finalizer interface{}) {
	c.Lock()
	defer c.Unlock()

	table, ok := c.table[k]
	if !ok {
		table = make([]*entry, 0, 10)
	}

	it := &entry{
		value:          x,
		timeExpiration: time.Now().Add(c.defaultExpiration),
	}
	if Finalizer != nil {
		if _, ok := Finalizer.(func()); ok {
			it.finalizer = Finalizer
		} else if _, ok := Finalizer.(func(x interface{})); ok {
			it.finalizer = Finalizer
		} else if _, ok := Finalizer.(func(x Value)); ok {
			it.finalizer = Finalizer
		} else {
			panic("Unknow Finalizer type!")
		}
	}

	table = append(table, it)
	c.table[k] = table
}

func (c *cache) Flush() {
	c.Lock()
	defer c.Unlock()

	changedItems := make(map[string][]*entry)
	for k, table := range c.table {
		for i := 0; i < len(table); i++ {
			if table[i].Expired() {
				table[i].Finalize()
				table[i], table = table[len(table)-1], table[:len(table)-1]
				changedItems[k] = table
				i--
			}
		}
	}
	for k, table := range changedItems {
		if len(table) == 0 {
			delete(c.table, k)
		} else {
			c.table[k] = table
		}
	}
}

func (c *cache) Clean() {
	c.Lock()
	defer c.Unlock()

	for _, table := range c.table {
		for i := 0; i < len(table); i++ {
			table[i].Finalize()
		}
	}
	c.table = nil
}

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) run(c *cache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			c.Flush()
		case <-j.stop:
			c.Clean()
			return
		}
	}
}

func startJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
	}
	c.janitor = j
	go j.run(c)
}

func stopJanitor(c **cache) {
	(*c).janitor.stop <- true
}
