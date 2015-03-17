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
	null bool
	f    ScoreFunc
}

// Score implements the Viterbier interface.
func (v vvalue) Score(n int) float64 {
	return v.f(n)
}

// IsNull implements the Viterbier interface.
func (v vvalue) IsNull() bool {
	return v.null
}

// Create a simple graph.
func simpleGraph() (*Graph, []interface{}) {

	// Some sequence of probabilities. (rows correspond to states.)
	scores := [][]float64{
		{0.1, 0.1, 0.2, 0.4, 0.11, 0.11, 0.12, 0.14},
		{0.4, 0.1, 0.3, 0.5, 0.21, 0.01, 0.12, 0.08},
		{0.2, 0.2, 0.4, 0.5, 0.09, 0.11, 0.32, 0.444},
	}

	// Convert to log scores.
	var n []interface{}
	for i, v := range scores {
		for j, _ := range v {
			scores[i][j] = math.Log(scores[i][j])
		}
		n = append(n, i)
	}
	// Define score functions to return state probabilities.
	var s1Func = func(n interface{}) float64 {
		return scores[0][n.(int)]
	}
	var s2Func = func(n interface{}) float64 {
		return scores[1][n.(int)]
	}
	var s3Func = func(n interface{}) float64 {
		return scores[2][n.(int)]
	}
	var s5Func = func(n interface{}) float64 {
		return scores[2][n.(int)]
	}
	var finalFunc = func(n interface{}) float64 {
		return 0
	}

	// Creates a new graph.
	g := New()

	// Create some nodes and assign values.
	g.Set("s0", vvalue{null: true}) // initial state
	g.Set("s1", vvalue{f: s1Func})
	g.Set("s2", vvalue{f: s2Func})
	g.Set("s3", vvalue{f: s3Func})
	g.Set("s5", vvalue{f: s5Func})
	g.Set("s4", vvalue{f: finalFunc, null: true}) // final state

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
	return g, n
}

// This example shows how to run a Viterbi decoder on a simple graph.
func ExampleDecoder() {

	var e error

	// Create the graph.
	g, sc := simpleGraph()
	e = fmt.Errorf("simple graph:\n%s\n", g)
	if e != nil {
		panic(e)
	}

	// Create a decoder.
	dec, e := NewDecoder(g)
	if e != nil {
		panic(e)
	}

	// Find the optimnal sequence.
	token := dec.Decode(sc)

	// The token has the backtrace to find the optimal path.
	fmt.Printf("\n\n>>>> FINAL: %s\n", token)
}

// Output:
