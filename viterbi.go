// Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"bytes"
	"fmt"
	"math"
)

type Viterbier interface {
	// Computes score for a given node at observation sequence index n.
	ScoreFunction(n int, node *Node) float64
}

type ScoreFunc func(n int, node *Node) float64

type Token struct {
	// Accumulated score for this hypothesis.
	Score float64
	// The optimal node sequence.
	Node *Node
	// Backtrace
	BT *Token
	// Sequence index.
	Index int
}

// Finds the sequence of nodes in the graph that maximizes the score of
// a sequence of N observations using the Viterbi algorithm.
// (http://en.wikipedia.org/wiki/Viterbi_algorithm)
// The node values must be of type Tokener.
type Decoder struct {
	graph  *Graph
	start  *Node
	end    *Node
	active []*Token
}

// Creates a new Viterbi decoder.
func NewDecoder(g *Graph, start, end *Node) (d *Decoder, e error) {

	// Check that all values in graph are of type Token.
	e = g.checkViterbier()
	if e != nil {
		return
	}

	d = &Decoder{graph: g, start: start, end: end, active: make([]*Token, 0)}

	// Initialization. First active hypothesis for start node.
	t := &Token{
		Score: 0,
		Node:  start,
		BT:    nil,
		Index: -1,
	}
	d.active = append(d.active, t)

	return
}

// Decodes a sequence of N observations.
// Returns the Viterbi path and total score.
// The node values must be of type Tokener.
func (d *Decoder) Decode(N int) (token *Token) {

	for i := 0; i < N; i++ {
		d.Propagate(i)
	}

	// Get the results from the terminal node.
	//bt = d.end.value.(*Token).Backtrace
	//score = d.end.value.(*Token).Score

	return
}

// Propagates tokens from nodes to succesors.
// Keeps the tokens that maximizes the score.
func (d *Decoder) Propagate(n int) {

	fmt.Printf("\nPROPAGATE: %d\n\n", n)

	// Data structure to hold candidate hypothesis before choosing the most likely.
	data := make(map[*Node][]*Token)

	for _, t := range d.active {
		for node, w := range t.Node.succesors {
			fmt.Printf("TOKEN: %s, SUCC: %s\n", t, node.key)
			_, found := data[node]
			if !found {
				data[node] = make([]*Token, 0)
			}
			// Copy and update Token.
			f := node.value.(Viterbier).ScoreFunction // scoring function for this node.
			nt := &Token{
				Score: t.Score + w + f(n, node),
				Node:  node,
				BT:    t,
				Index: n,
			}
			data[node] = append(data[node], nt)
		}
	}

	// We have all the candidates for all nodes. Keep the most likely.
	// Remove others.
	active := make([]*Token, 0)
	for _, node := range d.graph.nodes {

		candidates := data[node]
		var best *Token = nil
		max := -math.MaxFloat64
		for _, t := range candidates {
			//fmt.Printf("NODE: %s, TOKEN: %s\n", key, t.BacktraceString())
			if t.Score > max {
				max = t.Score
				best = t
			}
		}
		if best != nil {
			active = append(active, best)
		}
	}

	// Replace list of active hypothesis.
	d.active = active

	printActive(active)
	return
}

func (g *Graph) checkViterbier() error {

	for _, v := range g.nodes {
		_, ok := v.value.(Viterbier)
		if !ok {
			return fmt.Errorf("Value in node [%s] must implement the Viterbier interface.", v.key)
		}
	}
	return nil
}

func printActive(active []*Token) {

	for k, v := range active {
		fmt.Printf("%4d: %s\n", k, v)
	}
}

// Returns a slice fo tokens in tslice.
func (t *Token) Backtrace(tslice []*Token) {

	if t.BT == nil {
		return
	}
	tslice = append(tslice, t.BT)
	t.BT.Backtrace(tslice)
}

// Returns the backtrace as a string with the sequence of node keys.
func (t *Token) BacktraceString() string {

	bt := make([]*Token, 0)
	t.Backtrace(bt)
	buf := new(bytes.Buffer)
	for _, v := range bt {
		buf.WriteString(v.Node.key + " ")
	}
	return buf.String()
}

// Returns a astring with token and backtrace information.
func (t *Token) String() string {
	return fmt.Sprintf("n: %4d, node: %10s, sc: %6.2f, bt: %s ", t.Index, t.Node.key, t.Score, t.BacktraceString())
}
