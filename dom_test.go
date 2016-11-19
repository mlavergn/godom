// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
	. "golog"
	"io/ioutil"
	"path"
	"runtime"
	"testing"
)

func TestDOMSetContents(t *testing.T) {
	x := NewDOM()
	x.SetContents("<html></html>")
	if x.ContentLength() != 13 {
		t.Errorf("ContentLength %d vs expected %d", x.ContentLength(), 13)
	}
}

func loadData(t *testing.T, fileName string) (contents string) {
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
	contents := loadData(t, "test_a.html")

	d := NewDOM()
	d.SetContents(contents)

	meta := d.Find("meta", nil)
	if meta == nil {
		t.Errorf("failed to find META")
	}
}

func TestAttr(t *testing.T) {
	contents := loadData(t, "test_b.html")

	d := NewDOM()
	d.SetContents(contents)

	args := map[string]string{}

	node := d.Find("form", map[string]string{"id": "example_connect"})
	if node != nil {
		url := node[0].Attr("action")
		if len(url) > 0 {
			node = d.Find("input", map[string]string{"name": "cmd"})
			args["cmd"] = node[0].Attr("value")

			node = d.Find("input", map[string]string{"name": "user"})
			args["user"] = node[0].Attr("value")

			node = d.Find("input", map[string]string{"name": "password"})
			args["password"] = node[0].Attr("value")

			node = d.Find("input", map[string]string{"name": "url"})
			args["url"] = node[0].Attr("value")

			if args["user"] != "AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA@private" {
				d.Dump()
				t.Error("failed to parse arguments")
			}
		} else {
			d.Dump()
			t.Error("failed to find ACTION")
		}
	} else {
		d.Dump()
		t.Error("failed to find FORM")
	}
}

func TestChild(t *testing.T) {
	SetLogLevel(LOG_DEBUG)
	d := NewDOM()
	d.SetContents("<html><option id='2'>LOL</option><form action ='/foo'><select id='s'><option id='1'>Foo</option><option id='2'>Bar</option></select></form><div id='d'>Hello</div></html>")
	p := d.Find("select", map[string]string{"id": "s"})
	if len(p) != 1 {
		d.Dump()
		t.Errorf("failed to find SELECT node")
	}
	c := d.ChildFind(p[0], "option", map[string]string{"id": "2"})
	if len(c) != 1 || c[0].Attributes["id"] != "2" {
		d.Dump()
		t.Errorf("failed to find OPTION node")
	}
}

func TestDivText(t *testing.T) {
	SetLogLevel(LOG_DEBUG)
	d := NewDOM()
	d.SetContents("<html><div id=\"a\">Hello <strong>there</strong> world</div><div id=\"b\">Foo</div></html>")
	p := d.Find("div", map[string]string{"id": "a"})
	if len(p) != 1 {
		d.Dump()
		t.Errorf("failed to find node")
	} else {
		if p[0].Text() != "Hello world" {
			d.Dump()
			t.Errorf("failed to recombine text [%s]", p[0].Text)
		}
	}
}

func TestReaderText(t *testing.T) {
	SetLogLevel(LOG_DEBUG)
	d := NewDOM()
	d.SetContents("<html><div id=\"a\">Hello <strong>there<p>foo<strong>bar</strong></p></strong><bold>there</bold> world</div><div id=\"b\">Foo</div></html>")
	p := d.Find("div", map[string]string{"id": "a"})
	if len(p) != 1 {
		d.Dump()
		t.Errorf("failed to find node")
	} else {
		if p[0].ReaderText() != "Hello there foo bar there world" {
			d.Dump()
			t.Errorf("failed to generate reader text [%s]", p[0].Text)
		}
	}

}
