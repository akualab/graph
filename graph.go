// Original work: Copyright (c) 2013 Alexander Willing, All rights reserved.
// Modified work: Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package graph implements a weighted, directed graph data structure.
// See https://en.wikipedia.org/wiki/Graph_(abstract_data_type) for more information.
package graph

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math"
	"sort"
)

type Graph struct {
	nodes map[string]*Node // A map of all the nodes in this graph, indexed by their key.
}

type Node struct {
	key       string
	value     interface{}       // the stored value
	succesors map[*Node]float64 // maps the succesor node to the weight of the connection to it
}

var (
	DuplicateKeyError = errors.New("cannot merge node because key already exists")
)

// Returns the map of succesors.
func (node *Node) GetSuccesors() map[*Node]float64 {
	if node == nil {
		return nil
	}

	succesors := node.succesors
	return succesors
}

// Returns the node's key.
func (node *Node) Key() string {
	if node == nil {
		return ""
	}

	key := node.key
	return key
}

// Returns the node's value.
func (node *Node) Value() interface{} {
	if node == nil {
		return nil
	}

	value := node.value
	return value
}

// Initializes a new graph.
func New() *Graph {
	return &Graph{
		nodes: map[string]*Node{},
	}
}

// Returns the number of nodes contained in the graph.
func (g *Graph) Len() int {
	return len(g.nodes)
}

// If there is no node with the specified key yet, Set creates a new node and stores the value.
// Else, Set updates the value, but leaves all connections intact.
// Returns the node.
func (g *Graph) Set(key string, value interface{}) *Node {

	v := g.get(key)

	// if no such node exists
	if v == nil {
		// create a new one
		v = &Node{
			key:       key,
			value:     value,
			succesors: map[*Node]float64{},
		}

		// and add it to the graph
		g.nodes[key] = v
		return v
	}

	v.value = value
	return v
}

// Deletes the node with the specified key. Returns false if key is invalid.
func (g *Graph) Delete(key string) bool {

	// get node in question
	v := g.get(key)
	if v == nil {
		return false
	}

	// remove node from slice,
	delete(g.nodes, key)

	// remove arcs from other nodes to the node we are removing.
	for _, otherNode := range g.nodes {
		for succesor, _ := range otherNode.succesors {
			delete(succesor.succesors, v)
		}
	}
	return true
}

// Returns a slice containing all nodes. The slice is empty if the graph contains no nodes.
func (g *Graph) GetAll() (all []*Node) {
	for _, v := range g.nodes {
		all = append(all, v)
	}

	return
}

// Returns the node with this key, or nil and an error if there is no node with this key.
func (g *Graph) Get(key string) (v *Node, err error) {
	v = g.get(key)

	if v == nil {
		err = errors.New("graph: invalid key")
	}

	return
}

// Internal function
func (g *Graph) get(key string) *Node {
	return g.nodes[key]
}

// Creates an arc between the nodes specified by the keys. Returns false if one or both of the keys are invalid.
// If there already is a connection, it is overwritten with the new arc weight.
func (g *Graph) Connect(from string, to string, weight float64) bool {

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	v.succesors[otherV] = weight

	// success
	return true
}

// Creates an arc to a target node. Returns false if the target node is nil.
// If there already is a connection, it is overwritten with the new weight.
func (node *Node) Connect(toNode *Node, weight float64) bool {

	if toNode == nil {
		return false
	}

	node.succesors[toNode] = weight

	// success
	return true
}

// Removes an arc connecting the two nodes. Returns false if one or both of the keys are invalid.
func (g *Graph) Disconnect(from string, to string) bool {

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	// delete the arc
	delete(v.succesors, otherV)

	return true
}

// Removes arc connecting to target node. Returns false if target node is nil.
func (node *Node) Disconnect(toNode *Node) bool {

	if toNode == nil {
		return false
	}

	delete(node.succesors, toNode)

	// success
	return true
}

// Returns true and the arc weight if there is an arc between the nodes specified by their keys.
// Returns false if one or both keys are invalid or if there is no arc between the nodes.
func (g *Graph) IsConnected(from string, to string) (exists bool, weight float64) {

	fromV := g.get(from)
	if fromV == nil {
		return
	}

	toV := g.get(to)
	if toV == nil {
		return
	}

	// iterate over it's map of arcs; when the right node is found, return
	for succV, weight := range fromV.succesors {
		if succV == toV {
			return true, weight
		}
	}

	return
}

// Returns true and the arc weight if there is an arc to the target node.
// Returns false if there is no arc.
func (node *Node) IsConnected(toNode *Node) (exists bool, weight float64) {

	// iterate over it's map of arcs; when the right node is found, return
	for succV, weight := range node.succesors {
		if succV == toNode {
			return true, weight
		}
	}

	return
}

// Returns an identical copy of the graph.
func (g *Graph) Clone() (newG *Graph, e error) {

	// encode
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	e = enc.Encode(g)
	if e != nil {
		return
	}

	// now decode into new graph
	dec := gob.NewDecoder(buf)
	newG = New()
	e = dec.Decode(newG)

	return
}

// Converts arc weights to probabilities.
func (node *Node) Normalize(isLog bool) {

	var sum float64
	if !isLog {
		for _, w := range node.succesors {
			sum += w
		}
		for snode, w := range node.succesors {
			node.succesors[snode] = w / sum
		}
		return
	}

	// IsLog == true
	// convert to linear.
	for snode, w := range node.succesors {
		node.succesors[snode] = math.Exp(w)
	}
	// Normalize to probs. in linear domain.
	node.Normalize(false)

	// Convert to log probs.
	node.ConvertToLogProbs()
}

// For all nodes in graph, converts arc weights to probabilities.
func (g *Graph) Normalize(isLog bool) {

	for _, node := range g.nodes {
		node.Normalize(isLog)
	}
}

// Converts arc weights in linear domain to log weights.
func (node *Node) ConvertToLogProbs() {

	for snode, w := range node.succesors {
		node.succesors[snode] = math.Log(w)
	}
}

// For all nodes in graph, converts arc weights in linear domain to log weights.
func (g *Graph) ConvertToLogProbs() {

	for _, node := range g.nodes {
		node.ConvertToLogProbs()
	}
}

// Returns a slice of keys sorted alphabetically and the corresponding
// transition matrix of type [][]float64. Rows with no arcs
// have a nil slice. If isLog is true, missing connections are set to -Inf,
// zero otherwise.
func (g *Graph) TransitionMatrix(isLog bool) (keys []string, weights [][]float64) {

	n := g.Len()
	weights = make([][]float64, n)

	// Put nodes in a slice.
	nodes := make([]*Node, n)
	keys = make([]string, n)
	index := make(map[*Node]int)
	var k int
	for _, x := range g.nodes {
		nodes[k] = x
		k += 1
	}

	// Sort nodes by name.
	sort.Sort(ByName{nodes})

	// Map Node name to matrix index.
	for k, v := range nodes {
		index[v] = k
	}

	// Put transition weights in matrix.
	for _, fromNode := range nodes {
		i := index[fromNode]
		keys[i] = fromNode.key
		for toNode, w := range fromNode.succesors {
			j := index[toNode]
			if len(weights[i]) == 0 {
				weights[i] = make([]float64, n)
				if isLog {
					for m := 0; m < n; m++ {
						weights[i][m] = math.Inf(-1)
					}
				}
			}
			weights[i][j] = w
		}
	}
	return
}

// Merges copies of the graphs passed as a parameter.
// Nodes and arcs are copied to the new structure without modifying
// any connection. Returns DuplicateKeyError if any of the keys already exist.
func (g *Graph) Merge(graphs ...*Graph) error {

	// Verify that there are no duplicates before starting to merge.
	tmpMap := make(map[string]bool)
	for k, _ := range g.nodes {
		tmpMap[k] = true
	}
	for _, gg := range graphs {
		for k, _ := range gg.nodes {
			// Bail out if key already exists.
			_, found := tmpMap[k]
			if found {
				return DuplicateKeyError
			}
			tmpMap[k] = true
		}
	}

	// We are good, start merging.
	for _, gg := range graphs {

		// Copy graph.
		graph, e := gg.Clone()
		if e != nil {
			return e
		}
		for key, node := range graph.nodes {

			// Copy node to receiver.
			g.nodes[key] = node
		}
	}
	return nil
}

// Returns the graph as a string in YAML format.
func (g *Graph) String() (st string) {

	buf := new(bytes.Buffer)
	g.WriteYAML(buf)
	return buf.String()
}

// Sort Nodes.

type Nodes []*Node

func (s Nodes) Len() int      { return len(s) }
func (s Nodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByName implements sort.Interface by providing Less and using the Len and
type ByName struct{ Nodes }

func (s ByName) Less(i, j int) bool { return s.Nodes[i].key < s.Nodes[j].key }
