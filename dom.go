// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
  "fmt"
  "golang.org/x/net/html"
  "strings"
  "encoding/json"
)

// NodeAtributes map of strings keyed by strings
type NodeAttributes map[string]string
// JSONMap map of interface keyed by strings
type JSONMap map[string]interface{}

//
// DOM Node
//
type Node struct {
  tag string
  attributes NodeAttributes
  text string
}

//
// Node: Constructor.
//
func NewNode(tag string, attributes NodeAttributes) *Node {
  return &Node{tag:strings.ToLower(tag), attributes:attributes}
}

//
// Node: String representation.
//
func (self *Node) String() string {
  return fmt.Sprintf("Tag:\t%s\nAttr:\t%s\nText:\t%s\n", self.tag, self.attributes, self.text)
}

//
// Node: String with text contents.
//
func (self *Node) Text() string {
  return self.text
}

//
// Node: String with value of the provided attribute key.
//
func (self *Node) Attr(key string) string {
  return self.attributes[key]
}

//
// DOM Document.
//
type DOM struct {
  contents string
  document []*Node
}

//
// DOM: Constructor.
//
func NewDOM() *DOM {
  self := &DOM{}
  return self
}

//
// DOM: String representation.
//
func (self *DOM) String() string {
  result := ""
  for n := range self.document {
    result += fmt.Sprintf("Node:\n%s\n", self.document[n]) 
  }
  return result
}

//
// DOM: Parse the raw html contents.
//
func (self *DOM)SetContents(html string) {
  self.contents = html
  self._tokenize()
}

//
// DOM: The raw html contents.
//
func (self *DOM)Contents() string {
    return self.contents
}

//
// DOM: The byte count of the raw html contents.
//
func (self *DOM)ContentSize() int {
  return len(self.Contents())
}

//
// DOM: Parse the Token attributes into a map.
//
func (self *DOM)_tokenAttributes(t html.Token) NodeAttributes {
  attr := make(NodeAttributes)

  for _, a := range t.Attr {
    attr[a.Key] = a.Val
  }

  return attr
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
func (self *DOM)_tokenize() {
  z := html.NewTokenizer(strings.NewReader(self.contents))

  for {
    tt := z.Next()

    switch {
      case tt == html.ErrorToken:
        // end of document
        return
      case tt == html.StartTagToken:
        token := z.Token()
        self.document = append(self.document, NewNode(token.Data, self._tokenAttributes(token)))
        break
      case tt == html.EndTagToken:
        break
      case tt == html.TextToken:
        token := z.Token()
        dlen := len(self.document)
        if dlen >0 {
          node := self.document[dlen-1]
          if node != nil {
            node.text = strings.Trim(string(token.Data), " \r\n")
          }
        }
        break
    }
  }
}

//
//
//
func (self *DOM)DumpLinks() []string {
  result := make([]string, 100)
  for i := range self.document {
    node := self.document[i]
    if node.tag == "a" {
      attr := node.attributes
      value := attr["href"]
      if len(value) > 0 {
        result = append(result, value)
      }
    }
  }

  return result
}

//
// DOM: Find the Node of type tag with the specified attributes
//
func (self *DOM)Find(tag string, attributes NodeAttributes) *Node {
  var result *Node = nil

  var found bool
  var node *Node

  for i := range self.document {
    node = self.document[i]
    if node.tag == tag {
      // found a matching tag
      attributes := node.attributes
      found = true
      for k, v := range(attributes) {
        if attributes[k] != v {
          found = false
        }
      }
      if found {
        break
      }
    }
  }

  if found {
    result = node
  }

  return result
}

//
// DOM: Find the Node of type tag with text containing key
//
func (self *DOM)FindWithKey(tag string, substring string) *Node {
  var result *Node = nil

  var found bool
  var node *Node

  for i := range self.document {
    node = self.document[i]
    if node.tag == tag {
      // found a matching tag
      contents := node.text
      idx := strings.Index(contents, substring)
      if idx >= 0 {
        found = true
      }
      if found {
        break
      }
    }
  }

  if found {
    result = node
  }

  return result
}

//
// DOM: Find the given tag with the specified attributes
//
func (self *DOM)FindTextForClass(tag string, class string) string {
  result := "--"

  node := self.Find(tag, NodeAttributes{"class": class})

  if node != nil {
    result = node.text
  }

  return result
}

//
// DOM: Find the JSON key with text containing substring
//
func (self *DOM)FindJsonForScriptWithKey(substring string) JSONMap {
  var result JSONMap = nil

  node := self.FindWithKey("script", substring)

  if node != nil {
    contents := node.text
    idx := strings.Index(contents, substring)
    sub := contents[idx:]
    idx = strings.Index(sub, "}")
    if idx >= 0 {
      sub = sub[:idx+1]
    }

    // unmarshall is strict and wants complete JSON structures
    if !strings.HasPrefix(sub, "{") {
      sub = "{" + sub + "}"
    }

    bytes := []byte(sub)
    json.Unmarshal(bytes, &result)
  }

  return result
}
