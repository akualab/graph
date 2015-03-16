// Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"bytes"
	"fmt"
	"math"

	"github.com/golang/glog"
)

// The Viterbier interface is used to implement a Viterbi decoder using a directed graph.
// All that is needed to search the graph is to implement this interface in every node.
//
// For example, define a type "nodeValue" for the node values as follows:
//
//   type nodeValue struct {
//      // Plug your scoring function here.
//      f ScoreFunc
//   }
//
// and the method ScoreFunc which implements the Viterbier interface:
//
//   func (v nodeValue) ScoreFunc(n int) float64 {
//      return v.f(n)
//   }
//
type Viterbier interface {
	// Scoring function.
	Score(obs interface{}) (score float64)
}

// ScoreFunc is the scoring function type.
type ScoreFunc func(obs interface{}) float64

// Token is used to implement the token-passing algorithm.
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

// Decoder finds the sequence of nodes in the graph that maximizes
// the score of a sequence of N observations using the Viterbi algorithm.
// (see http://en.wikipedia.org/wiki/Viterbi_algorithm)
// The node values must be of type Token.
type Decoder struct {
	graph  *Graph
	start  *Node
	end    *Node
	active []*Token
}

// NewDecoder creates a new Viterbi decoder.
func NewDecoder(g *Graph, start, end *Node) (d *Decoder, e error) {

	// Check that all values in graph are of type Token.
	e = g.checkViterbier()
	if e != nil {
		return
	}

	d = &Decoder{graph: g, start: start, end: end, active: []*Token{}}

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

// Decode returns the Viterbi path and total score.
// The node values must be of type Viterbier.
func (d *Decoder) Decode(N int) *Token {

	for i := 0; i < N; i++ {
		d.Propagate(i)
	}

	var best *Token
	max := -math.MaxFloat64
	for _, t := range d.active {
		if t.Score > max {
			max = t.Score
			best = t
		}
	}
	return best
}

// Propagate tokens from nodes to successors.
// Keeps the tokens that maximizes the score.
func (d *Decoder) Propagate(n int) {

	glog.V(3).Infof("propagate: %d", n)

	// Data structure to hold candidate hypothesis before choosing the most likely.
	data := make(map[*Node][]*Token)

	for _, t := range d.active {
		for node, w := range t.Node.successors {
			glog.V(3).Infof("node:  %s, token: %s", node.key, t)
			_, found := data[node]
			if !found {
				data[node] = []*Token{}
			}
			// Copy and update Token.
			f := node.value.(Viterbier).Score // scoring function for this node.
			nt := &Token{
				Score: t.Score + w + f(n),
				Node:  node,
				BT:    t,
				Index: n,
			}
			data[node] = append(data[node], nt)
		}
	}

	// We have all the candidates for all nodes. Keep the most likely.
	// Remove others.
	var active []*Token
	for _, node := range d.graph.nodes {

		candidates := data[node]
		var best *Token
		max := -math.MaxFloat64
		for _, t := range candidates {
			if t.Score > max {
				max = t.Score
				best = t
			}
		}
		if best != nil {
			active = append(active, best)
		}
	}

	// Replace list of active hypotheses.
	d.active = active

	if glog.V(3) {
		glog.Info("active list:")
		printActive(active)
	}
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

// Backtrace returns a slice of tokens.
func (t *Token) Backtrace(tokens []*Token) []*Token {

	if t.BT == nil {
		return tokens
	}
	//tokens = append(tokens, t.BT)
	tokens = append(tokens, t)
	tokens = t.BT.Backtrace(tokens)
	return tokens
}

// BacktraceString returns the backtrace as a string
// with the sequence of node keys.
func (t *Token) BacktraceString() string {

	var bt []*Token
	bt = t.Backtrace(bt)

	buf := new(bytes.Buffer)
	_, err := buf.WriteString("| ")
	if err != nil {
		panic(err)
	}
	for i, _ := range bt {
		v := bt[len(bt)-i-1]
		st := fmt.Sprintf("%d:%s:%.2f | ", v.Index, v.Node.key, v.Score)
		_, err := buf.WriteString(st)
		if err != nil {
			panic(err)
		}

	}
	return buf.String()
}

// String returns a string with token and backtrace information.
func (t *Token) String() string {
	return fmt.Sprintf("n: %2d, node: %4s, sc: %4.2f, bt: %s ",
		t.Index, t.Node.key, t.Score, t.BacktraceString())
}
