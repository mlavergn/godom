// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// NodeAtributes map of strings keyed by strings
type NodeAttributes map[string]string

// JSONMap map of interface keyed by strings
type JSONMap map[string]interface{}

//
// DOM Node
//
type DOMNode struct {
	tag        string
	attributes NodeAttributes
	text       string
}

//
// Node: Constructor.
//
func NewDOMNode(tag string, attributes NodeAttributes) *DOMNode {
	return &DOMNode{tag: strings.ToLower(tag), attributes: attributes}
}

//
// Node: String representation.
//
func (self *DOMNode) String() string {
	return fmt.Sprintf("Tag:\t%s\nAttr:\t%s\nText:\t%s\n", self.tag, self.attributes, self.text)
}

//
// Node: String with text contents.
//
func (self *DOMNode) Text() string {
	return self.text
}

//
// Node: String with value of the provided attribute key.
//
func (self *DOMNode) Attr(key string) string {
	return self.attributes[key]
}

//
// DOM Document.
//
type DOM struct {
	contents string
	document []*DOMNode
	nodes    map[string][]*DOMNode
}

//
// DOM: Constructor.
//
func NewDOM() *DOM {
	self := &DOM{}
	self.nodes = make(map[string][]*DOMNode)
	return self
}

//
// DOM: String representation.
//
func (self *DOM) String() string {
	result := ""
	for _, node := range self.document {
		result += fmt.Sprintf("Node:\n%s\n", node)
	}
	return result
}

//
// DOM: Parse the raw html contents.
//
func (self *DOM) SetContents(html string) {
	self.contents = html
	self._parse(html)
}

//
// DOM: The raw html contents.
//
func (self *DOM) Contents() string {
	return self.contents
}

//
// DOM: The byte count of the raw html contents.
//
func (self *DOM) ContentSize() int {
	return len(self.Contents())
}

//
// DOM: Parse the Token attributes into a map.
//
func (self *DOM) _nodeAttributes(node *html.Node) (attrs NodeAttributes) {
	attrs = make(NodeAttributes)

	for _, attr := range node.Attr {
		attrs[attr.Key] = attr.Val
	}

	return attrs
}

//
// DOM: Parse the Token attributes into a map.
//
func (self *DOM) _parseFragment(root *html.Node, contents string) {
	nodes, err := html.ParseFragment(strings.NewReader(contents), root)
	if err == nil {
		for _, node := range nodes {
			self._walk(node, true)
		}
	}
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
func (self *DOM) _walk(node *html.Node, fragment bool) {
	parseSkipTags := map[string]int{"script": 1, "style": 1, "body": 1}
	fragmentSkipTags := map[string]int{"html": 1, "head": 1, "body": 1}

	switch node.Type {
	case html.ElementNode:
		if !fragment || (fragment && fragmentSkipTags[node.Data] == 0) {
			domNode := NewDOMNode(node.Data, self._nodeAttributes(node))
			self.document = append(self.document, domNode)
			nodeArr := self.nodes[domNode.tag]
			if nodeArr != nil {
				self.nodes[domNode.tag] = append(nodeArr, domNode)
			} else {
				self.nodes[domNode.tag] = []*DOMNode{domNode}
			}
		}
	case html.TextNode:
		text := strings.TrimSpace(node.Data)
		if len(text) > 0 {
			if node.Parent == nil || parseSkipTags[node.Parent.Data] == 0 {
				self._parseFragment(node.Parent, text)
			} else {
				dlen := len(self.document)
				if dlen > 0 {
					owningNode := self.document[dlen-1]
					if owningNode != nil {
						owningNode.text = strings.TrimSpace(node.Data)
					}
				}
			}
		}
	case html.CommentNode:
		domNode := NewDOMNode("comment", self._nodeAttributes(node))
		self.document = append(self.document, domNode)
	case html.ErrorNode:
		domNode := NewDOMNode("error", self._nodeAttributes(node))
		self.document = append(self.document, domNode)
	case html.DocumentNode:
		domNode := NewDOMNode("document", self._nodeAttributes(node))
		self.document = append(self.document, domNode)
	case html.DoctypeNode:
		domNode := NewDOMNode("doctype", self._nodeAttributes(node))
		self.document = append(self.document, domNode)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		self._walk(child, fragment)
	}
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
func (self *DOM) _parse(contents string) {
	doc, err := html.Parse(strings.NewReader(contents))
	if err == nil {
		self._walk(doc, false)
	}
}

//
//
//
func (self *DOM) DumpLinks() (result []string) {
	result = make([]string, 100)

	tagNodes := self.nodes["a"]
	for _, node := range tagNodes {
		attr := node.attributes
		value := attr["href"]
		if len(value) > 0 {
			result = append(result, value)
		}
	}

	return result
}

//
// DOM: Find the Node of type tag with the specified attributes
//
func (self *DOM) Find(tag string, attributes NodeAttributes) (result *DOMNode) {
	var found bool

	tagNodes := self.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		nodeAttrs := node.attributes
		found = true
		for k, v := range attributes {
			if nodeAttrs[k] != v {
				found = false
			}
		}
		if found {
			result = node
			break
		}
	}

	return result
}

//
// DOM: Find the Node of type tag with text containing key
//
func (self *DOM) FindWithKey(tag string, substring string) (result *DOMNode) {
	tagNodes := self.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		contents := node.text
		idx := strings.Index(contents, substring)
		if idx >= 0 {
      result = node
      break
		}
	}

	return result
}

//
// DOM: Find the given tag with the specified attributes
//
func (self *DOM) FindTextForClass(tag string, class string) (result string) {
	node := self.Find(tag, NodeAttributes{"class": class})

	if node != nil {
		result = node.text
	}

	return result
}

//
// DOM: Find the JSON key with text containing substring
//
func (self *DOM) FindJsonForScriptWithKey(substring string) (result JSONMap) {
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
