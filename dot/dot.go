// Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implements the gographviz interface: https://code.google.com/p/gographviz/
//
// Parses a dot-formatted graph like this:
//
//  digraph G {
//    x -> 2 [ label = 5.1 ];
//	  4 -> 2 [ label = 1 ];
//    4 -> x [ label = 2 ];
//    x -> x [ label = 0.3 ];
//  }
// where x, 2, 4 are the node keys and label = {5.1,1,2,0.3} are the weights.
package dot

import (
	"fmt"
	"strconv"

	graphviz "code.google.com/p/gographviz"
	"github.com/akualab/graph"
)

type GraphDOT struct {
	graph *graph.Graph
}

func NewGraphDOT() *GraphDOT {

	gd := new(GraphDOT)
	gd.graph = graph.New()
	return gd
}

func (gd *GraphDOT) SetStrict(strict bool) {}
func (gd *GraphDOT) SetDir(directed bool)  {}
func (gd *GraphDOT) SetName(name string)   {}

func (gd *GraphDOT) AddEdge(src, srcPort, dst, dstPort string, directed bool, attrs map[string]string) {

	w, err := strconv.ParseFloat(attrs["label"], 64)
	if err != nil {
		panic(err)
	}

	gd.graph.Set(src, nil)
	gd.graph.Set(dst, nil)
	ok := gd.graph.Connect(src, dst, w)
	if !ok {
		panic("Failed to connect.")
	}
}

func (gd *GraphDOT) AddNode(parentGraph string, name string, attrs map[string]string) {}
func (gd *GraphDOT) AddAttr(parentGraph string, field, value string)                  {}
func (gd *GraphDOT) AddSubGraph(parentGraph string, name string, attrs map[string]string) {
}

// Returns a *graph.Graph struct.
func (gd *GraphDOT) Graph() *graph.Graph {
	return gd.graph
}

// Converts a Graph to a string in DOT format.
// TODO: include node values.
func DOT(g *graph.Graph, name string) string {

	gv := graphviz.NewGraph()

	for _, node := range g.GetAll() {
		src := node.Key()
		gv.AddNode(name, src, nil)

		for succ, weight := range node.GetSuccesors() {
			dst := succ.Key()
			sw := map[string]string{"label": fmt.Sprintf("%f", weight)}
			gv.AddEdge(src, "", dst, "", true, sw)
		}
	}

	return gv.String()
}
