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

// The Graph object.
type Graph struct {
	// A map of all the nodes in this graph, indexed by their key.
	nodes map[string]*Node
}

// The Node object.
type Node struct {
	key   string
	value interface{}
	// Maps the successor node to the weight of the connection to it.
	successors map[*Node]float64
}

var (
	// ErrDuplicateKey is an error to indicare a naming conflict.
	ErrDuplicateKey = errors.New("cannot merge node because key already exists")
)

// Successors returns the map of successors.
func (node *Node) Successors() map[*Node]float64 {
	if node == nil {
		return nil
	}

	successors := node.successors
	return successors
}

// Key returns the node's key.
func (node *Node) Key() string {
	if node == nil {
		return ""
	}

	key := node.key
	return key
}

// Value returns the node's value.
func (node *Node) Value() interface{} {
	if node == nil {
		return nil
	}

	value := node.value
	return value
}

// New creates a graph.
func New() *Graph {
	return &Graph{
		nodes: map[string]*Node{},
	}
}

// Len returns the number of nodes contained in the graph.
func (g *Graph) Len() int {
	return len(g.nodes)
}

// Set returns a new or updated node.
// If key doesn't exist, Set creates a new node with value.
// If node with key exists, Set updates the value, all connections
// are unchanged.
func (g *Graph) Set(key string, value interface{}) *Node {

	v := g.get(key)

	// if no such node exists
	if v == nil {
		// create a new one
		v = &Node{
			key:        key,
			value:      value,
			successors: map[*Node]float64{},
		}

		// and add it to the graph
		g.nodes[key] = v
		return v
	}

	v.value = value
	return v
}

// Delete node by key. Returns false if key is invalid.
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
		for successor, _ := range otherNode.successors {
			delete(successor.successors, v)
		}
	}
	return true
}

// GetAll returns a slice containing all nodes.
func (g *Graph) GetAll() (all []*Node) {
	for _, v := range g.nodes {
		all = append(all, v)
	}
	return
}

// Predecessors returns a slice with the nodes that connect
// to this node.
func (g *Graph) Predecessors(node *Node) []*Node {

	pred := make(map[*Node]bool)
	var res []*Node

	// Mark nodes that have predesessors.
	for _, n := range g.nodes {
		yes, _ := n.IsConnected(node)
		if yes {
			pred[n] = true
		}
	}
	for v, _ := range pred {
		res = append(res, v)
	}
	return res
}

// StartNodes returns a slice of start nodes.
// A start node is a node with no predescessors.
func (g *Graph) StartNodes() []*Node {

	var res []*Node

	// Find nodes that have predesessors.
	for _, node := range g.nodes {
		if len(g.Predecessors(node)) == 0 {
			res = append(res, node)
		}
	}
	return res
}

// EndNodes returns a slice of end nodes.
// An end node is a node with no successors.
func (g *Graph) EndNodes() []*Node {

	var res []*Node

	// Find nodes that have successors.
	for _, node := range g.nodes {
		if len(node.successors) == 0 {
			res = append(res, node)
		}
	}
	return res
}

// Get node by key, returns an error if there is no node for key.
func (g *Graph) Get(key string) (v *Node, err error) {
	v = g.get(key)

	if v == nil {
		err = errors.New("graph: invalid key")
	}

	return
}

// Internal function.
func (g *Graph) get(key string) *Node {
	return g.nodes[key]
}

// Connect creates an arc between the nodes specified by the keys "from" and "to.
// Returns false if one or both keys are invalid.
// If a connection exists, it is overwritten with the new arc weight.
func (g *Graph) Connect(from string, to string, weight float64) bool {

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	v.successors[otherV] = weight

	// success
	return true
}

// Connect creates an arc between the node and a target node "toNode".
// Returns false if the target node is nil.
// If a connection exists, it is overwritten with the new arc weight.
func (node *Node) Connect(toNode *Node, weight float64) bool {

	if toNode == nil {
		return false
	}

	node.successors[toNode] = weight

	// success
	return true
}

// Disconnect removes an arc connecting the two nodes.
// Returns false if one or both of the keys are invalid.
func (g *Graph) Disconnect(from string, to string) bool {

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	// delete the arc
	delete(v.successors, otherV)

	return true
}

// Disconnect removes the arc fron node to "toNode".
// Returns false if target node is nil.
func (node *Node) Disconnect(toNode *Node) bool {

	if toNode == nil {
		return false
	}

	delete(node.successors, toNode)

	// success
	return true
}

// IsConnected returns true and the arc weight if arc exists.
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
	for succV, weight := range fromV.successors {
		if succV == toV {
			return true, weight
		}
	}
	return
}

// IsConnected returns true and the arc weight if arc exists.
// Returns false if there is no arc.
func (node *Node) IsConnected(toNode *Node) (exists bool, weight float64) {

	// iterate over it's map of arcs; when the right node is found, return
	for succV, weight := range node.successors {
		if succV == toNode {
			return true, weight
		}
	}
	return
}

// Clone returns a deep copy of the graph.
// The entire graph is serialized using the gob package.
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

// Normalize converts outbound arc weights to probabilities.
func (node *Node) Normalize(isLog bool) {
	var sum float64
	if !isLog {
		for _, w := range node.successors {
			sum += w
		}
		for snode, w := range node.successors {
			node.successors[snode] = w / sum
		}
		return
	}

	// IsLog == true
	// convert to linear.
	for snode, w := range node.successors {
		node.successors[snode] = math.Exp(w)
	}
	// Normalize to probs. in linear domain.
	node.Normalize(false)

	// Convert to log probs.
	node.ConvertToLogProbs()
}

// Normalize converts arc weights to probabilities such that the sum of
// the weights of the outbound arcs for a given node equals one.
func (g *Graph) Normalize(isLog bool) {
	for _, node := range g.nodes {
		node.Normalize(isLog)
	}
}

// ConvertToLogProbs converts arc weights to log probabilities.
func (node *Node) ConvertToLogProbs() {
	for snode, w := range node.successors {
		node.successors[snode] = math.Log(w)
	}
}

// ConvertToLogProbs converts arc weights to log probabilities.
func (g *Graph) ConvertToLogProbs() {
	for _, node := range g.nodes {
		node.ConvertToLogProbs()
	}
}

// TransitionMatrix returns a slice of keys sorted alphabetically and the corresponding
// transition matrix of type [][]float64. Rows with no outbound arcs
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
		for toNode, w := range fromNode.successors {
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

// Merge combines graphs as follows:
// Nodes and arcs are [deep] copied to the new structure without modifications.
// Returns ErrDuplicateKey if any of the keys is duplicated.
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
				return ErrDuplicateKey
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

// Add adds graphs as follows:
// Nodes and arcs are moved (not copied) to the main graph.
// Returns ErrDuplicateKey if any of the keys is duplicated.
// Both the main and added graphs will point to the same node and arc objects.
func (g *Graph) Add(graphs ...*Graph) error {

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
				return ErrDuplicateKey
			}
			tmpMap[k] = true
		}
	}

	// We are good, start merging.
	for _, gg := range graphs {

		// Add nodes to new graph
		for key, node := range gg.nodes {

			// Add node to main graph.
			g.nodes[key] = node
		}
	}
	return nil
}

// String returns the graph as a string in YAML format.
func (g *Graph) String() (st string) {

	buf := new(bytes.Buffer)
	err := g.WriteYAML(buf)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// Sort Nodes.

// Nodes is a slice of nodes.
type Nodes []*Node

// Len is the number of nodes to sort.
func (s Nodes) Len() int { return len(s) }

// Swap for sorting nodes.
func (s Nodes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByName implements the sort interface.
type ByName struct{ Nodes }

// Less implements the sort interface.
func (s ByName) Less(i, j int) bool { return s.Nodes[i].key < s.Nodes[j].key }
