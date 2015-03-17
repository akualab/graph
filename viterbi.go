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
//      null bool
//   }
//
// implements the Viterbier interface as follows:
//
//   func (v nodeValue) Score(o interface{}) float64 {
//      return v.f(o)
//   }
//
//   func (v nodeValue) IsNull() bool {
//      return v.null
//   }
//
type Viterbier interface {
	// Scoring function.
	Score(obs interface{}) (score float64)
	IsNull() bool
}

// ScoreFunc is the type of the scoring function.
type ScoreFunc func(obs interface{}) float64

// Token is used to implement the token-passing algorithm.
type Token struct {
	// Accumulated score for this hypothesis.
	Score float64
	// The optimal node sequence.
	Node *Node
	// Backtrace, list of linked tokens.
	BT *Token
	// Sequence index.
	Index int
}

// Decoder finds the sequence of nodes in the graph that maximizes
// the score of a sequence of N observations using the Viterbi algorithm.
// (see http://en.wikipedia.org/wiki/Viterbi_algorithm)
// The node values must implement the Viterbier interface.
type Decoder struct {
	graph  *Graph
	start  *Node
	end    *Node
	active []*Token
	hyps   map[*Node][]*Token
}

// NewDecoder creates a new Viterbi decoder.
// Graph must have exactly one start and one end node. Will return error otherwise.
func NewDecoder(g *Graph) (*Decoder, error) {

	// Search for start and end nodes.
	starts := g.StartNodes()
	if len(starts) != 1 {
		return nil, fmt.Errorf("graph must have exactly one start node. Found: %d", len(starts))
	}
	ends := g.EndNodes()
	if len(ends) != 1 {
		return nil, fmt.Errorf("graph must have exactly one end node. Found: %d", len(ends))
	}

	// Check that all values in graph implement the Viterbier interface.
	e := g.checkViterbier()
	if e != nil {
		return nil, e
	}

	d := &Decoder{graph: g, start: starts[0], end: ends[0], active: []*Token{}}

	// Initialization. First active hypothesis for start node.
	t := &Token{
		Score: 0,
		Node:  starts[0],
		BT:    nil,
		Index: -1,
	}
	d.active = append(d.active, t)

	return d, nil
}

// Decode returns the Viterbi path and total score.
// The argument is a slice of observations.
func (d *Decoder) Decode(obs []interface{}) *Token {

	for k, o := range obs {
		glog.V(5).Infof("propagate obs with index: %4d, value: %+v", k, o)
		d.propagate(k, o)
	}

	var best *Token
	max := math.Inf(-1)
	for _, t := range d.active {
		if t.Score > max {
			max = t.Score
			best = t
		}
	}
	return best
}

func (d *Decoder) newToken(prev *Token, node *Node, idx int, score float64) *Token {

	nt := &Token{
		Score: score,
		Node:  node,
		BT:    prev,
		Index: idx,
	}

	// No null nodes except for end node.
	if !node.Value().(Viterbier).IsNull() || node == d.end {
		d.hyps[node] = append(d.hyps[node], nt)
	}
	return nt
}

func (d *Decoder) pass(t *Token, idx int, o interface{}) {

	for node, w := range t.Node.successors {
		val := node.Value().(Viterbier)
		glog.V(6).Infof("pass from [%s] to [%s] null:%t, token: [%+v]", t.Node.key, node.key, val.IsNull(), t)

		// Keep passing recursively until finding an emitting node.
		if val.IsNull() {
			nt := d.newToken(t, node, idx, t.Score+w)
			glog.V(6).Infof("null node: %s, token: [%+v]", node.key, nt)
			d.pass(nt, idx, o)
		} else {
			f := node.value.(Viterbier).Score // scoring function for this node.
			nt := d.newToken(t, node, idx, t.Score+w+f(o))
			glog.V(6).Infof("emit node: %s, token: [%+v]", node.key, nt)
		}
	}
}

// Propagate tokens from nodes to successors.
// Keeps the tokens that maximizes the score.
func (d *Decoder) propagate(idx int, o interface{}) {

	// Init data structure to hold candidate hypothesis before choosing the most likely.
	// TODO consider avoid realloc memory
	d.hyps = make(map[*Node][]*Token)
	for _, node := range d.graph.GetAll() {
		d.hyps[node] = []*Token{}
	}

	// Iterate.
	for _, t := range d.active {
		d.pass(t, idx, o)
	}

	// We have all the candidates for all nodes. Keep the most likely.
	// Remove others.
	var active []*Token
	for _, node := range d.graph.nodes {

		candidates := d.hyps[node]
		var best *Token
		max := math.Inf(-1)
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

	if glog.V(6) {
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
		glog.Infof("active:%4d bt:%s", k, v)
	}
}

// Backtrace returns the Viterbi backtrace as an ordered
// slice of tokens.
func (t *Token) Backtrace(tokens []*Token) []*Token {

	if t.BT == nil {
		tokens = append(tokens, t)
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
	for i, _ := range bt {
		v := bt[len(bt)-i-1]
		st := fmt.Sprintf("{%d,%s,%.2f},", v.Index, v.Node.key, v.Score)
		_, err := buf.WriteString(st)
		if err != nil {
			panic(err)
		}
	}
	return buf.String()
}

// String returns a string with token and backtrace information.
func (t *Token) String() string {
	return fmt.Sprintf("n: %2d, node: %4s, sc: %4.2f, bt: {%s} ",
		t.Index, t.Node.key, t.Score, t.BacktraceString())
}
