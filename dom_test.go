// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
  "testing"
)

func TestDOMSetContents(t *testing.T) {
  x := NewDOM()
  x.SetContents("<html></html>")
  if x.ContentSize() != 13 {
    t.Errorf("ContentSize %d vs expected %d", x.ContentSize(), 13)
  }
}