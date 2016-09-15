// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
  "testing"
  "runtime"
  "path"
  "io/ioutil"
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
  contentPath := path.Join(path.Dir(filename), "_testdata", fileName)
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
}

func TestAttr(t *testing.T) {
  contents := _loadData(t, "test_b.html")

  d := NewDOM()
  d.SetContents(contents)

  args := map[string]string{}
  
  node := d.Find("form", map[string]string{"id":"example_connect"})
  if node != nil {
    url := node[0].Attr("action")
    if len(url) > 0 {
      node = d.Find("input", map[string]string{"name":"cmd"})
      args["cmd"] = node[0].Attr("value")

      node = d.Find("input", map[string]string{"name":"user"})
      args["user"] = node[0].Attr("value")

      node = d.Find("input", map[string]string{"name":"password"})
      args["password"] = node[0].Attr("value")

      node = d.Find("input", map[string]string{"name":"url"})
      args["url"] = node[0].Attr("value")

      if args["user"] != "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA@private" {
        t.Error("failed to parse arguments")
      }
    } else {
      t.Error("failed to find ACTION")
    }
  } else {
    t.Error("failed to find FORM")
  }
}
