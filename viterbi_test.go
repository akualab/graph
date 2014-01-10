// Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"math"
	"testing"
)

type vvalue struct {
	f ScoreFunc
}

func (v vvalue) ScoreFunction(n int, node *Node) float64 {
	return v.f(n, node)
}

func threeStateGraph(t *testing.T) *Graph {

	obs := [][]float64{{0.1, 0.1, 0.2, 0.4}, {0.4, 0.1, 0.3, 0.5}, {0.2, 0.2, 0.4, 0.5}}
	for i, v := range obs {
		for j, _ := range v {
			obs[i][j] = math.Log(obs[i][j])
		}
	}
	// Define score functions to implement edit distance.
	var s1Func = func(n int, node *Node) float64 {
		return obs[0][n]
	}
	var s2Func = func(n int, node *Node) float64 {
		return obs[1][n]
	}
	var s3Func = func(n int, node *Node) float64 {
		return obs[2][n]
	}
	var finalFunc = func(n int, node *Node) float64 {
		return 0
	}

	g := New()

	// set some nodes
	g.Set("s0", vvalue{}) // initial state
	g.Set("s1", vvalue{s1Func})
	g.Set("s2", vvalue{s2Func})
	g.Set("s3", vvalue{s3Func})
	g.Set("s4", vvalue{finalFunc}) // final state

	// make some connections
	g.Connect("s0", "s1", 1)
	g.Connect("s1", "s1", 0.4)
	g.Connect("s1", "s2", 0.5)
	g.Connect("s1", "s3", 0.1)
	g.Connect("s2", "s2", 0.3)
	g.Connect("s2", "s3", 0.7)
	g.Connect("s3", "s3", 0.4)
	g.Connect("s3", "s4", 0.6)

	g.ConvertToLogProbs()
	return g
}

func TestViterbi(t *testing.T) {

	var start, end *Node
	var e error

	g := threeStateGraph(t)
	t.Logf("three state graph:\n%s\n", g)

	if start, e = g.Get("s0"); e != nil {
		t.Fatal(e)
	}
	if end, e = g.Get("s4"); e != nil {
		t.Fatal(e)
	}

	dec, e := NewDecoder(g, start, end)
	if e != nil {
		t.Fatal(e)
	}
	dec.Decode(4)

}

// type Token struct {
// 	// Computes score for a given node at observation sequence index n.
// 	ScoreFunc ScoreFunc `yaml:"-" json:"-"`
// 	// Accumulated score for this hypothesis.
// 	Score float64
// 	// The optimal node sequence.
// 	Backtrace []*Node
// 	// A stored value.
// 	Value interface{}
// }
