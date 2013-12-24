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
	"errors"
	"sync"
)

type Node struct {
	key       string
	value     interface{}       // the stored value
	succesors map[*Node]float64 // maps the succesor node to the weight of the connection to it
	sync.RWMutex
}

// Returns the map of succesors.
func (v *Node) GetSuccesors() map[*Node]float64 {
	if v == nil {
		return nil
	}

	v.RLock()
	succesors := v.succesors
	v.RUnlock()

	return succesors
}

// Returns the nodees key.
func (v *Node) Key() string {
	if v == nil {
		return ""
	}

	v.RLock()
	key := v.key
	v.RUnlock()

	return key
}

// Returns the Nodes value.
func (v *Node) Value() interface{} {
	if v == nil {
		return nil
	}

	v.RLock()
	value := v.value
	v.RUnlock()

	return value
}

type Graph struct {
	nodes map[string]*Node // A map of all the nodes in this graph, indexed by their key.
	sync.RWMutex
}

// Initializes a new graph.
func New() *Graph {
	return &Graph{map[string]*Node{}, sync.RWMutex{}}
}

// Returns the amount of nodes contained in the graph.
func (g *Graph) Len() int {
	return len(g.nodes)
}

// If there is no node with the specified key yet, Set creates a new node and stores the value. Else, Set updates the value, but leaves all connections intact.
func (g *Graph) Set(key string, value interface{}) {
	// lock graph until this method is finished to prevent changes made by other goroutines
	g.Lock()
	defer g.Unlock()

	v := g.get(key)

	// if no such node exists
	if v == nil {
		// create a new one
		v = &Node{key, value, map[*Node]float64{}, sync.RWMutex{}}

		// and add it to the graph
		g.nodes[key] = v

		return
	}

	// else, just update the value
	v.Lock()
	v.value = value
	v.Unlock()
}

// Deletes the node with the specified key. Returns false if key is invalid.
func (g *Graph) Delete(key string) bool {
	// lock graph until this method is finished to prevent changes made by other goroutines while this one is looping etc.
	g.Lock()
	defer g.Unlock()

	// get node in question
	v := g.get(key)
	if v == nil {
		return false
	}

	// iterate over succesors, remove arcs from succesoring nodes
	for succesor, _ := range v.succesors {
		// delete arc to the to-be-deleted node
		succesor.Lock()
		delete(succesor.succesors, v)
		succesor.Unlock()
	}

	// delete node
	delete(g.nodes, key)

	return true
}

// Returns a slice containing all nodes. The slice is empty if the graph contains no nodes.
func (g *Graph) GetAll() (all []*Node) {
	g.RLock()
	for _, v := range g.nodes {
		all = append(all, v)
	}
	g.RUnlock()

	return
}

// Returns the node with this key, or nil and an error if there is no node with this key.
func (g *Graph) Get(key string) (v *Node, err error) {
	g.RLock()
	v = g.get(key)
	g.RUnlock()

	if v == nil {
		err = errors.New("graph: invalid key")
	}

	return
}

// Internal function, does NOT lock the graph, should only be used in between RLock() and RUnlock() (or Lock() and Unlock()).
func (g *Graph) get(key string) *Node {
	return g.nodes[key]
}

// Creates an arc between the nodes specified by the keys. Returns false if one or both of the keys are invalid.
// If there already is a connection, it is overwritten with the new arc weight.
func (g *Graph) Connect(from string, to string, weight float64) bool {

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	// add arc to node
	v.Lock()
	otherV.Lock()

	v.succesors[otherV] = weight

	v.Unlock()
	otherV.Unlock()

	// success
	return true
}

// Removes an arc connecting the two nodes. Returns false if one or both of the keys are invalid.
func (g *Graph) Disconnect(from string, to string) bool {

	// lock graph for reading until this method is finished to prevent changes made by other goroutines while this one is running
	g.RLock()
	defer g.RUnlock()

	// get nodes and check for validity of keys
	v := g.get(from)
	otherV := g.get(to)

	if v == nil || otherV == nil {
		return false
	}

	// delete the arc
	v.Lock()
	//otherV.Lock()

	delete(v.succesors, otherV)

	v.Unlock()
	//otherV.Unlock()

	return true
}

// Returns true and the arc weight if there is an arc between the nodes specified by their keys. Returns false if one or both keys are invalid, if they are the same, or if there is no arc between the nodes.
func (g *Graph) Adjacent(key string, otherKey string) (exists bool, weight float64) {
	// sanity check
	if key == otherKey {
		return
	}

	g.RLock()

	v := g.get(key)
	if v == nil {
		g.RUnlock()
		return
	}

	otherV := g.get(otherKey)
	if otherV == nil {
		g.RUnlock()
		return
	}

	g.RUnlock()

	v.RLock()
	defer v.RUnlock()
	otherV.RUnlock()
	defer otherV.RLock()

	// choose node with less arcs (easier to find 1 in 10 than to find 1 in 100)
	if len(v.succesors) < len(otherV.succesors) {
		// iterate over it's map of arcs; when the right node is found, return
		for iteratingV, weight := range v.succesors {
			if iteratingV == otherV {
				return true, weight
			}
		}
	} else {
		// iterate over it's map of arcs; when the right node is found, return
		for iteratingV, weight := range otherV.succesors {
			if iteratingV == v {
				return true, weight
			}
		}
	}

	return
}

// Returns the graph as a string in YAML format.
func (g *Graph) String() (st string) {

	buf := new(bytes.Buffer)
	g.WriteYAML(buf)
	return buf.String()
}
