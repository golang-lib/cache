// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"testing"
)

type IntCacheValue struct {
	size int
}

func (cv *IntCacheValue) Size() int {
	return cv.size
}

type SliceCacheValue []byte

func (mv SliceCacheValue) Size() int {
	return cap(mv)
}

func TestLRUInitialState(t *testing.T) {
	cache := NewLRUCache(5)
	l, sz, c, _ := cache.Stats()
	if l != 0 {
		t.Errorf("length = %v, want 0", l)
	}
	if sz != 0 {
		t.Errorf("size = %v, want 0", sz)
	}
	if c != 5 {
		t.Errorf("capacity = %v, want 5", c)
	}
}

func TestLRUSetInsertsValue(t *testing.T) {
	cache := NewLRUCache(100)
	data := &IntCacheValue{0}
	key := "key"
	cache.Set(key, data, 0)

	v, ok := cache.Get(key)
	if !ok || v.(*IntCacheValue) != data {
		t.Errorf("Cache has incorrect value: %v != %v", data, v)
	}
}

func TestLRUGetValueWithMultipleTypes(t *testing.T) {
	cache := NewLRUCache(100)
	data := &IntCacheValue{0}
	key := "key"
	cache.Set(key, data, 0)

	v, ok := cache.Get("key")
	if !ok || v.(*IntCacheValue) != data {
		t.Errorf("Cache has incorrect value for \"key\": %v != %v", data, v)
	}

	v, ok = cache.Get(string([]byte{'k', 'e', 'y'}))
	if !ok || v.(*IntCacheValue) != data {
		t.Errorf("Cache has incorrect value for []byte {'k','e','y'}: %v != %v", data, v)
	}
}

func TestLRUSetUpdatesSize(t *testing.T) {
	cache := NewLRUCache(100)
	emptyValue := &IntCacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue, 0)
	if _, sz, _, _ := cache.Stats(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0", sz)
	}
	someValue := &IntCacheValue{20}
	key = "key2"
	cache.Set(key, someValue, 0)
	if _, sz, _, _ := cache.Stats(); sz != 20 {
		t.Errorf("cache.Size() = %v, expected 20", sz)
	}
}

func TestLRUSetWithOldKeyUpdatesValue(t *testing.T) {
	cache := NewLRUCache(100)
	emptyValue := &IntCacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue, 0)
	someValue := &IntCacheValue{20}
	cache.Set(key, someValue, 0)

	v, ok := cache.Get(key)
	if !ok || v.(*IntCacheValue) != someValue {
		t.Errorf("Cache has incorrect value: %v != %v", someValue, v)
	}
}

func TestLRUSetWithOldKeyUpdatesSize(t *testing.T) {
	cache := NewLRUCache(100)
	emptyValue := &IntCacheValue{0}
	key := "key1"
	cache.Set(key, emptyValue, 0)

	if _, sz, _, _ := cache.Stats(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected %v", sz, 0)
	}

	someValue := &IntCacheValue{20}
	cache.Set(key, someValue, 0)
	expected := uint64(someValue.size)
	if _, sz, _, _ := cache.Stats(); sz != expected {
		t.Errorf("cache.Size() = %v, expected %v", sz, expected)
	}
}

func TestLRUGetNonExistent(t *testing.T) {
	cache := NewLRUCache(100)

	if _, ok := cache.Get("crap"); ok {
		t.Error("Cache returned a crap value after no inserts.")
	}
}

func TestLRUTake(t *testing.T) {
	cache := NewLRUCache(100)
	value := &IntCacheValue{1}
	key := "key"

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value, 0)

	// first take
	v, ok := cache.Take(key)
	if !ok || v.(*IntCacheValue) != value {
		t.Errorf("Cache has incorrect value: %v != %v", value, v)
	}

	// try again
	if _, ok = cache.Take(key); ok {
		t.Error("Cache returned a value after take.")
	}

	if _, sz, _, _ := cache.Stats(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0", sz)
	}

	if _, ok := cache.Get(key); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestLRUDelete(t *testing.T) {
	cache := NewLRUCache(100)
	value := &IntCacheValue{1}
	key := "key"

	if cache.Delete(key) {
		t.Error("Item unexpectedly already in cache.")
	}

	cache.Set(key, value, 0)

	if !cache.Delete(key) {
		t.Error("Expected item to be in cache.")
	}

	if _, sz, _, _ := cache.Stats(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0", sz)
	}

	if _, ok := cache.Get(key); ok {
		t.Error("Cache returned a value after deletion.")
	}
}

func TestLRUClear(t *testing.T) {
	cache := NewLRUCache(100)
	value := &IntCacheValue{1}
	key := "key"

	cache.Set(key, value, 0)
	cache.Clear()

	if _, sz, _, _ := cache.Stats(); sz != 0 {
		t.Errorf("cache.Size() = %v, expected 0 after Clear()", sz)
	}
}

func TestLRUCapacityIsObeyed(t *testing.T) {
	size := uint64(3)
	cache := NewLRUCache(size)
	value := &IntCacheValue{1}

	// Insert up to the cache's capacity.
	cache.Set("key1", value, 0)
	cache.Set("key2", value, 0)
	cache.Set("key3", value, 0)
	if _, sz, _, _ := cache.Stats(); sz != size {
		t.Errorf("cache.Size() = %v, expected %v", sz, size)
	}
	// Insert one more; something should be evicted to make room.
	cache.Set("key4", value, 0)
	if _, sz, _, _ := cache.Stats(); sz != size {
		t.Errorf("post-evict cache.Size() = %v, expected %v", sz, size)
	}
}

func TestLRUIsEvicted(t *testing.T) {
	size := uint64(3)
	cache := NewLRUCache(size)

	cache.Set("key1", &IntCacheValue{1}, 0)
	cache.Set("key2", &IntCacheValue{1}, 0)
	cache.Set("key3", &IntCacheValue{1}, 0)
	// lru: [key3, key2, key1]

	// Look up the elements. This will rearrange the LRU ordering.
	cache.Get("key3")
	cache.Get("key2")
	cache.Get("key1")
	// lru: [key1, key2, key3]

	cache.Set("key0", &IntCacheValue{1}, 0)
	// lru: [key0, key1, key2]

	// The least recently used one should have been evicted.
	if _, ok := cache.Get("key3"); ok {
		t.Error("Least recently used element was not evicted.")
	}
}

func BenchmarkLRUGet(b *testing.B) {
	cache := NewLRUCache(64 * 1024 * 1024)
	value := make(SliceCacheValue, 1000)
	cache.Set("stuff", value, 0)
	for i := 0; i < b.N; i++ {
		val, ok := cache.Get("stuff")
		if !ok {
			panic("error")
		}
		_ = val
	}
}
