// Copyright 2016 Marc Lavergne <mlavergn@gmail.com>. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package godom

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"strings"
	"sync"
)

// DOMNodeAttributes map of strings keyed by strings
type DOMNodeAttributes map[string]string

// JSONMap map of interface keyed by strings
type JSONMap map[string]interface{}

// JSONDelimiter type
type JSONDelimiter [2]string

// JSONArrayDelimiter type
var JSONArrayDelimiter = JSONDelimiter{"[", "]"}

// JSONDictionaryDelimiter type
var JSONDictionaryDelimiter = JSONDelimiter{"{", "}"}

// DOMNode def
//
type DOMNode struct {
	Index         int
	Tag           string
	Attributes    DOMNodeAttributes
	TextFragments []string
	Parent        *DOMNode
	Children      []*DOMNode
}

//
// NewDOMNode constructor
//
func NewDOMNode(index int, parent *DOMNode, tag string, attributes DOMNodeAttributes) DOMNode {
	return DOMNode{
		Index:      index,
		Parent:     parent,
		Children:   []*DOMNode{},
		Tag:        strings.ToLower(tag),
		Attributes: attributes,
	}
}

//
// Node: String representation.
//
func (id *DOMNode) String() (desc string) {
	desc = ""
	desc += fmt.Sprintf("\nIndex:\t%d\nTag:\t%s\nAttr:\t%s\nTextFragments:\t%s\n", id.Index, id.Tag, id.Attributes, id.TextFragments)
	if id.Parent != nil {
		desc += fmt.Sprintf("Edges:\n\tParent:\t%d - %s\n", id.Parent.Index, id.Parent.Tag)
	}
	if len(id.Children) != 0 {
		for _, child := range id.Children {
			desc += fmt.Sprintf("\tChild:\t%d - %s\n", child.Index, child.Tag)
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

// Text export
func (id *DOMNode) Text() (result string) {
	// Join() has a 2x performance penalty over len() for single fragments
	// Single fragments comprise 99% of fragments, thus the switch/case
	switch len(id.TextFragments) {
	case 0:
	case 1:
		result = id.TextFragments[0]
	default:
		result = strings.Join(id.TextFragments, " ")
	}

	return
}

//
// ReaderText recombines the node text fragments into the human reader visibile text
//
func (id *DOMNode) ReaderText() (result string) {
	fragCount := len(id.TextFragments)
	childCount := len(id.Children)

	// recursively recombine any child node text fragments
	if childCount == 0 {
		// there are no children, defer to Text()
		result = id.Text()
	} else {
		i := 0

		// there may not be a leading fragment, be we assume there is
		if i < fragCount {
			result += id.TextFragments[i] + " "
			i++
		}

		// interleave child text and fragments until there are no more
		for _, child := range id.Children {
			childText := child.ReaderText()
			if len(childText) > 0 {
				result += childText + " "
				if fragCount > 2 && i < fragCount {
					result += id.TextFragments[i] + " "
					i++
				}
			}
		}
		result = strings.TrimSpace(result)

		// any trailing fragments, this should be at most count 1
		if i < fragCount {
			result += " " + strings.Join(id.TextFragments[i:], " ")
		}
	}

	return
}

//
// DOM Document.
//
type DOM struct {
	contents  string
	document  []*DOMNode
	nodes     map[string][]*DOMNode
	rootNode  *DOMNode
	nodeCount int
}

//
// NewDOM Constructor
//
func NewDOM() DOM {
	return DOM{
		nodes: map[string][]*DOMNode{},
	}
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
func (id *DOM) SetContents(htmlString string) {
	id.contents = htmlString

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		log.Println(err)
		return
	}
	id.parseHTMLNode(nil, doc, false)
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
	return len(id.contents)
}

//
// RootNode : The HTML root node
//
func (id *DOM) RootNode() (result *DOMNode) {
	if id.rootNode == nil {
		// we're looking for the tidy-ed HTML node at index 1
		// there's the childless DOCUMENT node at index 0
		for i := 0; i < len(id.document); i++ {
			if id.document[i].Tag == "html" {
				id.rootNode = id.document[i]
			}
		}
	}

	return id.rootNode
}

//
// Dump : dump the textual representation of the DOM
//
func (id *DOM) Dump() {
	log.Println(id.document)
}

//
// DOM: Parse the Token attributes into a map.
//
func (id *DOM) parseHTMLNodeAttributes(node *html.Node) (attrs DOMNodeAttributes) {
	attrs = make(DOMNodeAttributes)

	// NOTE: keys never have whitespace once parsed / values (even IDs) retain whitespace
	// parse the []html.Attribute into a hashmap
	for _, attr := range node.Attr {
		attrs[attr.Key] = attr.Val
	}

	return attrs
}

//
// DOM: Parse the Token attributes into a map.
//
func (id *DOM) parseHTMLFragment(parent *DOMNode, current *html.Node, contents string) {
	nodes, err := html.ParseFragment(strings.NewReader(contents), current)
	if err == nil {
		for _, node := range nodes {
			id.parseHTMLNode(parent, node, true)
		}
	}
}

//
// DOM: Walk the DOM and parse the HTML tokens into Nodes.
//
var (
	parseSkipTags    map[string]int
	fragmentSkipTags map[string]int
	once             sync.Once
)

func (id *DOM) parseHTMLNode(parent *DOMNode, current *html.Node, fragment bool) {
	// constant candidates
	once.Do(func() {
		parseSkipTags = map[string]int{"script": 1, "style": 1, "body": 1}
		fragmentSkipTags = map[string]int{"html": 1, "head": 1, "body": 1}
	})

	switch current.Type {
	case html.ElementNode:
		if !fragment || (fragment && fragmentSkipTags[current.Data] == 0) {
			id.nodeCount++
			domNode := NewDOMNode(id.nodeCount, parent, current.Data, id.parseHTMLNodeAttributes(current))
			// set the children and swap
			if parent != nil {
				parent.Children = append(parent.Children, &domNode)
			}
			parent = &domNode
			id.document = append(id.document, &domNode)
			nodeArr := id.nodes[domNode.Tag]
			if nodeArr != nil {
				id.nodes[domNode.Tag] = append(nodeArr, &domNode)
			} else {
				id.nodes[domNode.Tag] = []*DOMNode{&domNode}
			}
		}
	case html.TextNode:
		text := strings.TrimSpace(current.Data)
		if strings.Index(text, "<") != -1 && (current.Parent == nil || parseSkipTags[current.Parent.Data] == 0) {
			id.parseHTMLFragment(parent, current.Parent, text)
		} else {
			// we need to handle structures like (eg. <div>foo<strong>baz</strong>bar</div>)
			// Assumption: if the current node already has text, it belongs to the parent
			currentNode := id.document[len(id.document)-1]
			if currentNode != nil && len(currentNode.TextFragments) != 0 {
				currentNode = currentNode.Parent
			}
			if currentNode != nil {
				currentNode.TextFragments = append(currentNode.TextFragments, text)
			}
		}
	case html.CommentNode:
		id.nodeCount++
		domNode := NewDOMNode(id.nodeCount, parent, "comment", id.parseHTMLNodeAttributes(current))
		id.document = append(id.document, &domNode)
	case html.ErrorNode:
		id.nodeCount++
		domNode := NewDOMNode(id.nodeCount, parent, "error", id.parseHTMLNodeAttributes(current))
		id.document = append(id.document, &domNode)
	case html.DocumentNode:
		id.nodeCount++
		domNode := NewDOMNode(id.nodeCount, parent, "document", id.parseHTMLNodeAttributes(current))
		id.document = append(id.document, &domNode)
	case html.DoctypeNode:
		id.nodeCount++
		domNode := NewDOMNode(id.nodeCount, parent, "doctype", id.parseHTMLNodeAttributes(current))
		id.document = append(id.document, &domNode)
	}

	// recurse for all child nodes
	for child := current.FirstChild; child != nil; child = child.NextSibling {
		id.parseHTMLNode(parent, child, fragment)
	}
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
		rootNode := id.RootNode()
		for node != nil && parent.Index <= node.Index && node != rootNode {
			if node.Parent == parent {
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
	return id.ChildFind(id.RootNode(), tag, attributes)
}

//
// ChildFind : Find the child Node of type tag with the specified attributes
//
func (id *DOM) ChildFind(parent *DOMNode, tag string, attributes DOMNodeAttributes) (result []*DOMNode) {
	tagNodes := id.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		nodeAttrs := node.Attributes
		found := true
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
	return id.ChildFindWithKey(id.RootNode(), tag, substring)
}

//
// ChildFindWithKey : Find the child Node of type tag with text containing key
//
func (id *DOM) ChildFindWithKey(parent *DOMNode, tag string, substring string) (result []*DOMNode) {
	tagNodes := id.nodes[tag]
	for _, node := range tagNodes {
		// found a matching tag
		if id.IsDescendantNode(parent, node) {
			contents := node.Text()
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
	return id.ChildFindTextForClass(id.RootNode(), tag, class)
}

//
// ChildFindTextForClass : Find the child given tag with the specified attributes
//
func (id *DOM) ChildFindTextForClass(parent *DOMNode, tag string, class string) (result string) {
	nodes := id.ChildFind(parent, tag, DOMNodeAttributes{"class": class})

	if len(nodes) > 0 {
		result = nodes[0].Text()
	}

	return result
}

//
// FindJSONForScriptWithKey : Find the JSON key with text containing substring
//
func (id *DOM) FindJSONForScriptWithKey(substring string) (result JSONMap, err error) {
	return id.ChildFindJSONForScriptWithKeyDelimiter(id.RootNode(), substring, JSONDictionaryDelimiter)
}

//
// FindJSONForScriptWithKeyDelimiter : Find the JSON key with delimited text containing substring
//
func (id *DOM) FindJSONForScriptWithKeyDelimiter(substring string, delimiter JSONDelimiter) (result JSONMap, err error) {
	return id.ChildFindJSONForScriptWithKeyDelimiter(id.RootNode(), substring, delimiter)
}

//
// ChildFindJSONForScriptWithKeyDelimiter : Find the child JSON key with delimited text containing substring
//
func (id *DOM) ChildFindJSONForScriptWithKeyDelimiter(parent *DOMNode, substring string, delimiter JSONDelimiter) (result JSONMap, err error) {
	nodes := id.ChildFindWithKey(parent, "script", substring)

	if len(nodes) > 0 {
		contents := nodes[0].Text()
		idx := strings.Index(contents, substring)
		sub := contents[idx:]
		idx = strings.Index(sub, delimiter[1])
		if idx >= 0 {
			sub = sub[:idx+1]
		}

		// unmarshall is strict and wants complete JSON structures
		if !strings.HasPrefix(sub, delimiter[0]) {
			idx = strings.Index(sub, delimiter[0])
			if idx > 0 {
				sub = sub[idx:]
			} else {
				sub = delimiter[0] + sub + delimiter[1]
			}
		}

		// no newlines
		sub = strings.Replace(sub, "\n", "", -1)
		// no tabs
		sub = strings.Replace(sub, "\t", "", -1)

		bytes := []byte(sub)
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			if strings.HasPrefix(err.Error(), "invalid character ") {
				// JSON improper escaping detected - need to split the string and tidy it
				log.Println("Tidy JSON")
				subtidy := delimiter[0]
				entries := strings.Split(sub[1:len(sub)-1], ",")
				for _, entry := range entries {
					val := strings.Split(entry, ":")
					subtidy += fmt.Sprintf("\"%s\": \"%s\",", strings.Trim(val[0], " '"), strings.Trim(val[1], " '"))
				}
				subtidy = subtidy[:len(subtidy)-1] + delimiter[1]

				bytes := []byte(subtidy)
				err = json.Unmarshal(bytes, &result)
			}
		}

		// we may have reset err above, so recheck
		if err != nil {
			log.Println(err, "\n", sub)
		}
	}

	return
}
