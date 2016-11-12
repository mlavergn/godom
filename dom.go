// Copyright 2016, Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package godom

import (
	."golog"
	"golang.org/x/net/html"
	"encoding/json"
	"fmt"
	"strings"
	"strconv"
)

// DOMNodeAttributes map of strings keyed by strings
type DOMNodeAttributes map[string]string

// JSONMap map of interface keyed by strings
type JSONMap map[string]interface{}

// DOMNode def
//
type DOMNode struct {
	Index      int
	Tag        string
	Attributes DOMNodeAttributes
	Text       string
	Parent     *DOMNode
	Children   []*DOMNode
}

//
// NewDOMNode constructor
//
func NewDOMNode(parent *DOMNode, tag string, attributes DOMNodeAttributes) *DOMNode {
	return &DOMNode{Parent: parent, Children: []*DOMNode{}, Tag: strings.ToLower(tag), Attributes: attributes}
}

//
// Node: String representation.
//
func (id *DOMNode) String() (desc string) {
	desc = ""
	desc += fmt.Sprintf("\nIndex:\t%d\nTag:\t%s\nAttr:\t%s\nText:\t%s\n", id.Index, id.Tag, id.Attributes, id.Text)
	if id.Parent != nil {
		desc += "Parent:\t" + strconv.Itoa(id.Parent.Index) + "-" + id.Parent.Tag + "\n"
	}
	if len(id.Children) != 0 {
		for _, child := range id.Children {
			desc += "Child:\t" + child.Tag + "\n"
		}
	}
	return desc
}

//
// Attr Node: String with value of the provided attribute key.
//
func (id *DOMNode) Attr(key string) string {
	return id.Attributes[key]
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
// NewDOM Constructor
//
func NewDOM() *DOM {
	id := &DOM{}
	id.nodes = make(map[string][]*DOMNode)
	return id
}

//
// DOM: String representation.
//
func (id *DOM) String() (result string) {
	result = ""
	for _, node := range id.document {
		result += fmt.Sprintf("Node:\n%s\n", node)
	}
	return result
}

//
// SetContents : parse the raw html contents.
//
func (id *DOM) SetContents(html string) {
	id.contents = html
	id._parse(html)
}

//
// Contents : The raw html contents.
//
func (id *DOM) Contents() string {
	return id.contents
}

//
// ContentLength : The byte count of the raw html contents.
//
func (id *DOM) ContentLength() int {
	return len(id.Contents())
}

//
// RootNode : The HTML root node
//
var rootNode *DOMNode
func (id *DOM) RootNode() (result *DOMNode) {
	if rootNode == nil {
		// we're looking for the tidy-ed HTML node at index 1
		// there's the childless DOCUMENT node at index 0
		for i:=0; i<len(id.document); i++ {
			if id.document[i].Tag == "html" {
				rootNode = id.document[i]
			}
		}
	}

	return rootNode
}

//
// DumpNodes : The byte count of the raw html contents.
//
func (id *DOM) DumpNodes() {
		LogInfo(id.document)
}

func (id *DOM) DumpRelationships() {
	for _, node := range id.document {
		LogInfo(node.Parent)
	}
}

//
// DOM: Parse the Token attributes into a map.
//
func (id *DOM) _nodeAttributes(node *html.Node) (attrs DOMNodeAttributes) {
	attrs = make(DOMNodeAttributes)

	for _, attr := range node.Attr {
		attrs[attr.Key] = attr.Val
	}

	return attrs
}

//
// DOM: Parse the Token attributes into a map.
//
func (id *DOM) _parseFragment(parent *DOMNode, root *html.Node, contents string) {
	nodes, err := html.ParseFragment(strings.NewReader(contents), root)
	if err == nil {
		for _, node := range nodes {
			id._walk(parent, node, true)
		}
	}
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
func (id *DOM) _walk(parent *DOMNode, root *html.Node, fragment bool) {
	parseSkipTags := map[string]int{"script": 1, "style": 1, "body": 1}
	fragmentSkipTags := map[string]int{"html": 1, "head": 1, "body": 1}

	switch root.Type {
	case html.ElementNode:
		if !fragment || (fragment && fragmentSkipTags[root.Data] == 0) {
			domNode := NewDOMNode(parent, root.Data, id._nodeAttributes(root))
			// set the children and swap
			if parent != nil {
				parent.Children = append(parent.Children, domNode)
			}
			parent = domNode
			id.document = append(id.document, domNode)
			domNode.Index = len(id.document)
			nodeArr := id.nodes[domNode.Tag]
			if nodeArr != nil {
				id.nodes[domNode.Tag] = append(nodeArr, domNode)
			} else {
				id.nodes[domNode.Tag] = []*DOMNode{domNode}
			}
		}
	case html.TextNode:
		text := strings.TrimSpace(root.Data)
		if len(text) > 0 {
			if root.Parent == nil || parseSkipTags[root.Parent.Data] == 0 {
				id._parseFragment(parent, root.Parent, text)
			} else {
				dlen := len(id.document)
				if dlen > 0 {
					owningNode := id.document[dlen-1]
					if owningNode != nil {
						owningNode.Text = strings.TrimSpace(root.Data)
					}
				}
			}
		}
	case html.CommentNode:
		domNode := NewDOMNode(parent, "comment", id._nodeAttributes(root))
		id.document = append(id.document, domNode)
		domNode.Index = len(id.document)
	case html.ErrorNode:
		domNode := NewDOMNode(parent, "error", id._nodeAttributes(root))
		id.document = append(id.document, domNode)
		domNode.Index = len(id.document)
	case html.DocumentNode:
		domNode := NewDOMNode(parent, "document", id._nodeAttributes(root))
		id.document = append(id.document, domNode)
		domNode.Index = len(id.document)
	case html.DoctypeNode:
		domNode := NewDOMNode(parent, "doctype", id._nodeAttributes(root))
		id.document = append(id.document, domNode)
		domNode.Index = len(id.document)
	}

	for child := root.FirstChild; child != nil; child = child.NextSibling {
		id._walk(parent, child, fragment)
	}
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
func (id *DOM) _parse(contents string) {
	doc, err := html.Parse(strings.NewReader(contents))
	if err == nil {
		id._walk(nil, doc, false)
	}
}

//
// DumpLinks : Show the a hrefs in the DOM
//
func (id *DOM) DumpLinks() (result []string) {
	result = make([]string, 100)

	tagNodes := id.nodes["a"]
	for _, node := range tagNodes {
		attr := node.Attributes
		value := attr["href"]
		if len(value) > 0 {
			result = append(result, value)
		}
	}

	return result
}

//
// IsDescendantNode : Is node a descendant of parent?
// The fastest confirmation is bottom up since the relationships are
// one to many.
//
func (id *DOM) IsDescendantNode(parent *DOMNode, node *DOMNode) (result bool) {
	result = false
	// nil and equality check
	if parent == nil {
		result = true
	} else if node == nil {
		result = false
	} else if parent == node {
		result = true
	} else {
		// we would have matched above if parent and node were the root node
		root := id.RootNode()
		for node != nil && parent.Index <= node.Index && node.Index != root.Index {
			if node.Parent.Index == parent.Index {
				result = true
				break
			} else {
				node = node.Parent
			}
		}
	}

	return result
}

//
// IsChildNode : Is node a child of parent?
// The fastest confirmation is bottom up since the relationships are
// one to many.
//
func (id *DOM) IsChildNode(parent *DOMNode, node *DOMNode) (result bool) {
	result = false
	if node != nil && node.Parent == parent {
		result = true
	}

	return result
}

//
// Find : Find the Node of type tag with the specified attributes
//
func (id *DOM) Find(tag string, attributes DOMNodeAttributes) (result []*DOMNode) {
	return id.NodeFind(id.RootNode(), tag, attributes)
}

//
// NodeFind : Find the child Node of type tag with the specified attributes
//
func (id *DOM) NodeFind(parent *DOMNode, tag string, attributes DOMNodeAttributes) (result []*DOMNode) {
	found := true
	tagNodes := id.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		nodeAttrs := node.Attributes
		found = true
		for k, v := range attributes {
			if nodeAttrs[k] != v {
				found = false
			}
		}
		if found {
			if id.IsDescendantNode(parent, node) {
				result = append(result, node)
			}
		}
	}

	return result
}

//
// FindWithKey : Find the Node of type tag with text containing key
//
func (id *DOM) FindWithKey(tag string, substring string) (result []*DOMNode) {
	return id.NodeFindWithKey(id.RootNode(), tag, substring)
}

//
// NodeFindWithKey : Find the child Node of type tag with text containing key
//
func (id *DOM) NodeFindWithKey(parent *DOMNode, tag string, substring string) (result []*DOMNode) {
	tagNodes := id.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		if id.IsDescendantNode(parent, node) {
			contents := node.Text
			idx := strings.Index(contents, substring)
			if idx >= 0 {
				result = append(result, node)
			}
		}
	}

	return result
}

//
// FindTextForClass : Find the given tag with the specified attributes
//
func (id *DOM) FindTextForClass(tag string, class string) (result string) {
	return id.NodeFindTextForClass(id.RootNode(), tag, class)
}

//
// NodeFindTextForClass : Find the child given tag with the specified attributes
//
func (id *DOM) NodeFindTextForClass(parent *DOMNode, tag string, class string) (result string) {
	nodes := id.NodeFind(parent, tag, DOMNodeAttributes{"class": class})

	if len(nodes) > 0 {
		result = nodes[0].Text
	}

	return result
}

//
// FindJSONForScriptWithKey : Find the JSON key with text containing substring
//
func (id *DOM) FindJSONForScriptWithKey(substring string) (result JSONMap) {
	return id.NodeFindJSONForScriptWithKey(id.RootNode(), substring)
}

//
// NodeFindJSONForScriptWithKey : Find the child JSON key with text containing substring
//
func (id *DOM) NodeFindJSONForScriptWithKey(parent *DOMNode, substring string) (result JSONMap) {
	nodes := id.NodeFindWithKey(parent, "script", substring)

	if len(nodes) > 0 {
		contents := nodes[0].Text
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
