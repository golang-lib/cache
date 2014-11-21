// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type LRUCache struct {
	mu sync.Mutex

	// list & table of *entry objects
	list  *list.List
	table map[string]*list.Element

	// Our current size, in bytes. Obviously a gross simplification and low-grade
	// approximation.
	size uint64

	// How many bytes we are limiting the cache to.
	capacity uint64
}

func NewLRUCache(capacity uint64) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}

func (lru *LRUCache) Get(key string) (v interface{}, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}
	lru.moveToFront(element)
	return element.Value.(*entry).value, true
}

func (lru *LRUCache) Set(key string, value interface{}, size int) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	size = adjustValueSize(value, size)
	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value, size)
	} else {
		lru.addNew(key, value, size)
	}
}

func (lru *LRUCache) SetIfAbsent(key string, value interface{}, size int) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	size = adjustValueSize(value, size)
	if element := lru.table[key]; element != nil {
		lru.moveToFront(element)
	} else {
		lru.addNew(key, value, size)
	}
}

func (lru *LRUCache) Take(key string) (v interface{}, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= uint64(element.Value.(*entry).size)
	return element.Value.(*entry).value, true
}

func (lru *LRUCache) Delete(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= uint64(element.Value.(*entry).size)
	return true
}

func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
}

func (lru *LRUCache) SetCapacity(capacity uint64) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.capacity = capacity
	lru.checkCapacity()
}

func (lru *LRUCache) Stats() (length, size, capacity uint64, oldest time.Time) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).timeAccessed
	}
	return uint64(lru.list.Len()), lru.size, lru.capacity, oldest
}

func (lru *LRUCache) StatsJSON() string {
	if lru == nil {
		return "{}"
	}
	l, s, c, o := lru.Stats()
	return fmt.Sprintf("{\"Length\": %v, \"Size\": %v, \"Capacity\": %v, \"OldestAccess\": \"%v\"}", l, s, c, o)
}

func (lru *LRUCache) Keys() []string {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	keys := make([]string, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}

func (lru *LRUCache) Items() []Item {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	items := make([]Item, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(*entry)
		items = append(items, Item{Key: v.key, Value: v.value, Size: v.size})
	}
	return items
}

func (lru *LRUCache) updateInplace(element *list.Element, value interface{}, size int) {
	sizeDiff := size - element.Value.(*entry).size
	element.Value.(*entry).value = value
	element.Value.(*entry).size = size
	lru.size += uint64(sizeDiff)
	lru.moveToFront(element)
	lru.checkCapacity()
}

func (lru *LRUCache) moveToFront(element *list.Element) {
	lru.list.MoveToFront(element)
	element.Value.(*entry).timeAccessed = time.Now()
}

func (lru *LRUCache) addNew(key string, value interface{}, size int) {
	newEntry := &entry{
		key:          key,
		value:        value,
		size:         size,
		timeAccessed: time.Now(),
	}
	element := lru.list.PushFront(newEntry)
	lru.table[key] = element
	lru.size += uint64(newEntry.size)
	lru.checkCapacity()
}

func (lru *LRUCache) checkCapacity() {
	// Partially duplicated from Delete
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entry)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= uint64(delValue.size)
	}
}
