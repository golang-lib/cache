// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

func adjustValueSize(value interface{}, size int) int {
	if size <= 0 {
		if v, ok := value.(Value); ok {
			return v.Size()
		}
	}
	return size
}
