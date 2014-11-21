// Copyright 2013 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"testing"
)

func TestCache(t *testing.T) {
	tc := NewCache(0, 0)

	a, found := tc.Fetch("a")
	if found {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Fetch("b")
	if found {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Fetch("c")
	if found {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Store("a", 1, nil)
	tc.Store("b", "b", nil)
	tc.Store("c", 3.5, nil)

	x, found := tc.Fetch("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if x == nil {
		t.Error("x for a is nil")
	}
	if a2, ok := x.(int); !ok {
		t.Error("x is a int")
	} else if a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	if _, found := tc.Fetch("a"); found {
		t.Error("a was fetched!")
	}
}
