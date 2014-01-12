// Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"fmt"
	"math"
)

// Create a type to implement the Viterbier interface.
type vvalue struct {
	f ScoreFunc
}

// Implements the Viterbier interface.
func (v vvalue) ScoreFunction(n int, node *Node) float64 {
	return v.f(n, node)
}

// Create a simple graph.
func simpleGraph() *Graph {

	// Some sequence of probabilities. (rows correspond to states.)
	obs := [][]float64{
		{0.1, 0.1, 0.2, 0.4, 0.11, 0.11, 0.12, 0.14},
		{0.4, 0.1, 0.3, 0.5, 0.21, 0.01, 0.12, 0.08},
		{0.2, 0.2, 0.4, 0.5, 0.09, 0.11, 0.32, 0.444},
	}

	// Convert to log probabilities.
	for i, v := range obs {
		for j, _ := range v {
			obs[i][j] = math.Log(obs[i][j])
		}
	}
	// Define score functions to return state probabilities.
	var s1Func = func(n int, node *Node) float64 {
		return obs[0][n]
	}
	var s2Func = func(n int, node *Node) float64 {
		return obs[1][n]
	}
	var s3Func = func(n int, node *Node) float64 {
		return obs[2][n]
	}
	var s5Func = func(n int, node *Node) float64 {
		return obs[2][n]
	}
	var finalFunc = func(n int, node *Node) float64 {
		return 0
	}

	// Creates a new graph.
	g := New()

	// Create some nodes and assign values.
	g.Set("s0", vvalue{}) // initial state
	g.Set("s1", vvalue{s1Func})
	g.Set("s2", vvalue{s2Func})
	g.Set("s3", vvalue{s3Func})
	g.Set("s5", vvalue{s5Func})
	g.Set("s4", vvalue{finalFunc}) // final state

	// Make connections.
	g.Connect("s0", "s1", 1)
	g.Connect("s1", "s1", 0.4)
	g.Connect("s1", "s2", 0.5)
	g.Connect("s1", "s3", 0.1)
	g.Connect("s2", "s2", 0.5)
	g.Connect("s2", "s3", 0.4)
	g.Connect("s2", "s5", 0.1)
	g.Connect("s5", "s5", 0.7)
	g.Connect("s5", "s1", 0.3)
	g.Connect("s3", "s3", 0.6)
	g.Connect("s3", "s4", 0.4)

	// Convert transition probabilities to log.
	g.ConvertToLogProbs()
	return g
}

// This example shows how to run a Viterbi decoder on a simple graph.
func ExampleDecoder() {

	var start, end *Node
	var e error

	// Create the graph.
	g := simpleGraph()
	e = fmt.Errorf("simple graph:\n%s\n", g)
	panic(e)

	// Define the start and end nodes.
	if start, e = g.Get("s0"); e != nil {
		panic(e)
	}
	if end, e = g.Get("s4"); e != nil {
		panic(e)
	}

	// Create a decoder.
	dec, e := NewDecoder(g, start, end)
	if e != nil {
		panic(e)
	}

	// Find the optimnal sequence.
	token := dec.Decode(8)

	// The token has the backtrace to find the optimal path.
	fmt.Printf("\n\n>>>> FINAL: %s\n", token)
}

// Output:
