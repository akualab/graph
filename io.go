// Original work: Copyright (c) 2013 Alexander Willing, All rights reserved.
// Modified work: Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"launchpad.net/goyaml"
)

// Struct to export/import a graph.
type GraphIO struct {
	inv map[*Node]string
	// Node values indexed by key.
	Nodes map[string]interface{} `json:"nodes"`
	// Arc weight indexed by start node and end node keys.
	Arcs map[string]map[string]float64 `json:"arcs"`
}

// adds a key - node pair to the GraphIO
func (g GraphIO) add(v *Node) {
	// set the key - node pair
	g.Nodes[v.key] = v.value

	g.Arcs[v.key] = map[string]float64{}

	// for each successor...
	for successor, weight := range v.successors {
		// save the arc connection to the successor into the arcs map
		g.Arcs[v.key][successor.key] = weight
	}
}

// Prepares a graph for export.
func (g *Graph) exportGraph() (gio *GraphIO) {
	// build inverted map
	inv := map[*Node]string{}
	for key, v := range g.nodes {
		if _, ok := inv[v]; !ok {
			inv[v] = key
		}
	}

	gio = &GraphIO{inv, map[string]interface{}{}, map[string]map[string]float64{}}

	// add nodes and arcs to gio
	for _, v := range g.nodes {
		gio.add(v)
	}

	return
}

// Encodes the graph into a []byte. With this method, graph implements the
// gob.GobEncoder interface.
func (g *Graph) GobEncode() ([]byte, error) {
	// build inverted map
	inv := map[*Node]string{}
	for key, v := range g.nodes {
		if _, ok := inv[v]; !ok {
			inv[v] = key
		}
	}

	gGob := GraphIO{inv, map[string]interface{}{}, map[string]map[string]float64{}}

	// add nodes and arcs to gGob
	for _, v := range g.nodes {
		gGob.add(v)
	}

	// encode gGob
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(gGob)

	return buf.Bytes(), err
}

// Decodes a []byte into the graphs nodes and arcs. With this method, graph implements the
// gob.GobDecoder interface.
func (g *Graph) GobDecode(b []byte) (err error) {
	// decode into GraphIO
	gGob := &GraphIO{}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)

	err = dec.Decode(gGob)
	if err != nil {
		return
	}

	// set the nodes
	for key, value := range gGob.Nodes {
		g.Set(key, value)
	}

	// connect the nodes
	for key, successors := range gGob.Arcs {
		for otherKey, weight := range successors {
			if ok := g.Connect(key, otherKey, weight); !ok {
				return errors.New("invalid arc endpoints")
			}
		}
	}

	return
}

// Writes Graph to an io.Writer in YAML.
func (g *Graph) WriteYAML(w io.Writer) error {

	gio := g.exportGraph()
	b, err := goyaml.Marshal(gio)
	if err != nil {
		return err
	}
	_, e := w.Write(b)
	return e
}

// Implements json.Marshaler interface.
func (g *Graph) MarshalJSON() (b []byte, e error) {

	gio := g.exportGraph()
	b, e = json.Marshal(gio)

	return
}

// Implements json.Unmarshaler interface.
func (g *Graph) UnmarshalJSON(b []byte) error {

	gio := &GraphIO{}
	e := json.Unmarshal([]byte(b), gio)
	if e != nil {
		return e
	}

	e = gio.initGraph(g)
	if e != nil {
		return e
	}

	return nil

}

// Implements goyaml.Getter interface.
func (g *Graph) GetYAML() (tag string, value interface{}) {

	value = g.exportGraph()
	return
}

// Implements goyaml.Setter interface.
func (g *Graph) SetYAML(tag string, value interface{}) bool {

	// Not sure this is right. I need to get the byte slice before
	// unmarshaling into gio. The SetYAML method gives me a the object.
	// My solution is to marshal the value into bytes first.
	b, err := goyaml.Marshal(value)
	if err != nil {
		panic(err)
	}

	gio := &GraphIO{}
	err = goyaml.Unmarshal(b, gio)
	if err != nil {
		panic(err)
	}

	e := gio.initGraph(g)
	if e != nil {
		return false
	}

	return true
}

func (gio *GraphIO) initGraph(g *Graph) (e error) {

	// set the nodes
	for key, value := range gio.Nodes {
		g.Set(key, value)
	}

	// connect the nodes
	for key, successors := range gio.Arcs {
		for otherKey, weight := range successors {
			if ok := g.Connect(key, otherKey, weight); !ok {
				return errors.New("invalid arc endpoints")
			}
		}
	}

	return
}

// Reads graph in JSON format.
func ReadJSONGraph(fn string) (*Graph, error) {

	dat, e := ioutil.ReadFile(fn)
	if e != nil {
		return nil, e
	}

	g := New()
	e = json.Unmarshal(dat, g)
	if e != nil {
		return nil, e
	}

	return g, nil
}

// Write graph in JSON format.
func (g *Graph) WriteJSONGraph(fn string) error {

	b, e := g.MarshalJSON()
	if e != nil {
		return e
	}

	return ioutil.WriteFile(fn, b, 0644)
}
