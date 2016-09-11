// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
  "testing"
  "runtime"
  "path"
  "io/ioutil"
  "log"
)

func TestDOMSetContents(t *testing.T) {
  x := NewDOM()
  x.SetContents("<html></html>")
  if x.ContentSize() != 13 {
    t.Errorf("ContentSize %d vs expected %d", x.ContentSize(), 13)
  }
}

func _loadData(t *testing.T, fileName string) (contents string) {
  _, filename, _, _ := runtime.Caller(0)
  contentPath := path.Join(path.Dir(filename), "_testdata", "test_a.html")
  bytes, err := ioutil.ReadFile(contentPath)
  if err != nil {
    t.Errorf("%s", err)
  }

  contents = string(bytes)

  return contents
}

func TestFind(t *testing.T) {
  contents := _loadData(t, "test_a.html")

  d := NewDOM()
  d.SetContents(contents)
  
  meta := d.Find("meta", nil)
  if meta == nil {
    t.Errorf("failed to find META")
  }

  log.Println(meta)
}