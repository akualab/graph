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
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"launchpad.net/goyaml"
)

func TestConnect(t *testing.T) {

	g := sampleGraph(t)

	// test connections
	ok, weight := g.IsConnected("1", "2")
	if !ok || weight != 5 {
		t.Fail()
	}

	ok, weight = g.IsConnected("1", "3")
	if !ok || weight != 1 {
		t.Fail()
	}

	ok, weight = g.IsConnected("2", "3")
	if !ok || weight != 9 {
		t.Fail()
	}

	ok, weight = g.IsConnected("4", "2")
	if !ok || weight != 3 {
		t.Fail()
	}

	// test non-connections
	ok, _ = g.IsConnected("1", "4")
	if ok {
		t.Fail()
	}
}

func TestNodeConnect(t *testing.T) {

	var ok bool
	var weight float64

	g := sampleGraph(t)

	// get nodes
	node1, node2, node3, node4 := getSampleNodes(t, g)

	ok, weight = node1.IsConnected(node2)
	if !ok || weight != 5 {
		t.Fail()
	}

	ok, weight = node1.IsConnected(node3)
	if !ok || weight != 1 {
		t.Fail()
	}

	ok, weight = node2.IsConnected(node3)
	if !ok || weight != 9 {
		t.Fail()
	}

	ok, weight = node4.IsConnected(node2)
	if !ok || weight != 3 {
		t.Fail()
	}

	// test non-connections
	ok, _ = node1.IsConnected(node4)
	if ok {
		t.Fail()
	}

	// test disconnect
	ok = node1.Disconnect(node2)
	if !ok {
		t.Fatalf("Failed to disconnect.")
	}

	// create a new sample graph.
	g1 := sampleGraph(t)

	// They should NOT match.
	if e := compareGraphs(g, g1); e == nil {
		t.Fatalf("Graph matched, expected no match.")
	}

	// Reconnect the nodes.
	ok = node1.Connect(node2, 5)
	if !ok {
		t.Fatalf("Failed to connect.")
	}

	// They should match now.
	if e := compareGraphs(g, g1); e != nil {
		t.Fatal(e)
	}
}

func TestDelete(t *testing.T) {

	g := sampleGraph(t)

	// preserve a pointer to node "1"
	one := g.get("1")
	if one == nil {
		t.Fail()
	}

	// delete node
	ok := g.Delete("1")
	if !ok {
		t.Fail()
	}

	// make sure it's not in the graph anymore
	deletedOne := g.get("1")
	if deletedOne != nil {
		t.Fail()
	}

	// test for orphaned connections
	succ := g.get("2").GetSuccesors()
	for n, _ := range succ {
		if n == one {
			t.Fail()
		}
	}

	succ = g.get("3").GetSuccesors()
	for n, _ := range succ {
		if n == one {
			t.Fail()
		}
	}
}

func TestTransitionMatrix(t *testing.T) {

	g := sampleGraph(t)
	keys, weights := g.TransitionMatrix(false)
	lastKey := ""

	for i, from := range keys {

		// check alphabetic order.
		if keys[i] < lastKey {
			t.Fatalf("not in alphabetic order expected [%s] >=  [%s]", keys[i], lastKey)
		}

		// skip if no arcs!
		if len(weights[i]) == 0 {
			continue // no arcs for this node.
		}
		for j, to := range keys {
			ok, w := g.IsConnected(from, to)
			if ok {
				if weights[i][j] != w {
					t.Fatalf("weights don't match [%f] vs. [%f]", weights[i][j], w)
				}
			} else {
				if weights[i][j] != 0 {
					t.Fatalf("expected zero weight, got [%f]", weights[i][j])
				}
			}
		}
	}
}

func TestLogTransitionMatrix(t *testing.T) {

	g := sampleGraph(t)
	keys, weights := g.TransitionMatrix(true)

	for i, from := range keys {

		// skip if no arcs!
		if len(weights[i]) == 0 {
			continue // no arcs for this node.
		}
		for j, to := range keys {
			ok, w := g.IsConnected(from, to)
			if ok {
				if weights[i][j] != w {
					t.Fatalf("weights don't match [%f] vs. [%f]", weights[i][j], w)
				}
			} else {
				if weights[i][j] != math.Inf(-1) {
					t.Fatalf("expected zero weight, got [%f]", weights[i][j])
				}
			}
		}
	}
}

func TestLogProbs(t *testing.T) {

	g0 := sampleGraph(t)

	g1, _ := g0.Clone()
	g1.NormalizeWeights(true)

	keys, weights0 := g0.TransitionMatrix(false)
	_, weights1 := g1.TransitionMatrix(true)
	n := len(weights0)

	if n != len(weights1) {
		t.Fatalf("length mismatch [%d] vs. [%d]", n, len(weights1))
	}

	for i := 0; i < n; i++ {

		// skip if no arcs!
		if len(weights0[i]) == 0 {
			continue // no arcs for this node.
		}

		var sum float64
		for m := 0; m < n; m++ {
			sum += weights0[i][m]
		}

		for j := 0; j < n; j++ {
			ok0, _ := g0.IsConnected(keys[i], keys[j])
			ok1, _ := g1.IsConnected(keys[i], keys[j])
			t.Logf("i=%d, j=%d, conn=%v", i, j, ok0)
			if ok0 != ok1 {
				t.Fatalf("connection mismatch from [%s] to [%s]", keys[i], keys[j])
			}

			if ok0 {
				w0n := math.Log(weights0[i][j] / sum)
				t.Logf("weights [%f] vs. [%f]", w0n, weights1[i][j])
				if w0n != weights1[i][j] {
					t.Fatalf("weights don't match [%f] vs. [%f]", w0n, weights1[i][j])
				}
			} else {
				if weights0[i][j] != 0 {
					t.Fatalf("expected zero weight, got [%f]", weights0[i][j])
				}
				if weights1[i][j] != math.Inf(-1) {
					t.Fatalf("expected -Inf weight, got [%f]", weights1[i][j])
				}

			}
		}
	}
}

func TestClone(t *testing.T) {
	g := sampleGraph(t)

	g1, e := g.Clone()
	if e != nil {
		t.Fatal(e)
	}

	if e := compareGraphs(g, g1); e != nil {
		t.Fatal(e)
	}
}

func TestGob(t *testing.T) {
	g := sampleGraph(t)

	// encode
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	err := enc.Encode(g)
	if err != nil {
		fmt.Println(err)
	}

	// now decode into new graph
	dec := gob.NewDecoder(buf)
	newG := New()
	err = dec.Decode(newG)
	if err != nil {
		fmt.Println(err)
	}

	// validate length of new graph
	if len(g.nodes) != len(newG.nodes) {
		t.Fail()
	}

	// validate contents of new graph
	for k, v := range g.nodes {
		if newV := newG.get(k); newV.value != v.value {
			t.Fail()
		}
	}
}

func TestYAML(t *testing.T) {

	// Get the sample graph.
	g0 := sampleGraph(t)

	// Create sample graph JSON file in temp dir for testing.
	fn := os.TempDir() + "graph.yaml"
	t.Logf("yaml file: %s", fn)
	err := ioutil.WriteFile(fn, []byte(graphData), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Read YAML file.
	dat, ee := ioutil.ReadFile(fn)
	if ee != nil {
		t.Fatal(err)
	}

	g1 := New()
	err = goyaml.Unmarshal(dat, g1)
	if err != nil {
		panic(err)
	}

	if e := compareGraphs(g0, g1); e != nil {
		t.Fatal(e)
	}

	// Write YAML file.
	b, eb := goyaml.Marshal(g0)
	if eb != nil {
		panic(eb)
	}
	fn = os.TempDir() + "graph2.yaml"
	err = ioutil.WriteFile(fn, b, 0644)
	if err != nil {
		panic(err)
	}

}

func TestJSON(t *testing.T) {

	// Get the sample graph.
	g0 := sampleGraph(t)
	fn := os.TempDir() + "graph.json"
	t.Logf("json file: %s", fn)

	// Write YAML file.
	b, eb := json.Marshal(g0)
	if eb != nil {
		panic(eb)
	}
	err := ioutil.WriteFile(fn, b, 0644)
	if err != nil {
		panic(err)
	}

	// Read JSON file back and compare.
	dat, ee := ioutil.ReadFile(fn)
	if ee != nil {
		t.Fatal(err)
	}

	g2 := New()
	err = json.Unmarshal(dat, g2)
	if err != nil {
		panic(err)
	}

	if e := compareGraphs(g0, g2); e != nil {
		t.Fatal(e)
	}

}

func ExampleGraph() {
	g := New()

	// set key → value pairs
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// connect nodes/nodes
	g.Connect("1", "2", 5)
	g.Connect("1", "3", 1)
	g.Connect("2", "3", 9)
	g.Connect("4", "2", 3)

	// delete a node, and all connections to it
	g.Delete("1")

	// encode into buffer
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	err := enc.Encode(g)
	if err != nil {
		fmt.Println(err)
	}

	// now decode into new graph
	dec := gob.NewDecoder(buf)
	newG := New()
	err = dec.Decode(newG)
	if err != nil {
		fmt.Println(err)
	}
}

// Checks if there is a mismatch between two graphs.
// NOTE. The value in node is an interface. When unmarshaling, the value
// may be interpreted as int or float64. We convert int to float64 to
// compare numbers of the same type.
func compareGraphs(g1, g2 *Graph) (e error) {

	if e := includeGraphs(g1, g2); e != nil {
		return e
	}
	// reverse order.
	return includeGraphs(g2, g1)

}

// Checks if g2 is included in g1.
func includeGraphs(g1, g2 *Graph) (e error) {

	// check length
	if len(g1.nodes) != len(g2.nodes) {
		return fmt.Errorf("graph length mismatch")
	}

	// check node contents
	for k1, v1 := range g1.nodes {
		val, ok := (v1.value).(int)
		if ok {
			v1.value = float64(val)
		}

		v2 := g2.get(k1)
		val, ok = (v2.value).(int)
		if ok {
			v2.value = float64(val)
		}

		if v2.value != v1.value {
			return fmt.Errorf("graph content mismatch. [%+v] of type [%T] vs. [%+v] of type [%T]", v1.value, v1.value, v2.value, v2.value)
		}
	}

	// check connections.
	for k1, v1 := range g1.nodes {
		for v2, w1 := range v1.succesors {
			k2 := v2.key

			// check if there is connection from k1 to k2 in the other graph.
			ok, w2 := g2.IsConnected(k1, k2)
			if !ok {
				return fmt.Errorf("arc mismatch. from [%s] to [%s]", k1, k2)
			}
			if w1 != w2 {
				return fmt.Errorf("weight mismatch. from [%s] to [%s], got weight [%f] vs weight [%f]", w1, w2)
			}
		}
	}
	return
}

func sampleGraph(t *testing.T) *Graph {

	g := New()

	// set some nodes
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")

	// make some connections
	ok := g.Connect("1", "2", 5)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("1", "3", 1)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("2", "3", 9)
	if !ok {
		t.Fail()
	}

	ok = g.Connect("4", "2", 3)
	if !ok {
		t.Fail()
	}

	return g
}

func getSampleNodes(t *testing.T, g *Graph) (node1, node2, node3, node4 *Node) {

	var e error
	node1, e = g.Get("1")
	if e != nil {
		t.Fatal(e)
	}
	node2, e = g.Get("2")
	if e != nil {
		t.Fatal(e)
	}
	node3, e = g.Get("3")
	if e != nil {
		t.Fatal(e)
	}
	node4, e = g.Get("4")
	if e != nil {
		t.Fatal(e)
	}
	return
}

func printNodes(vSlice map[string]*Node) {
	for _, v := range vSlice {
		fmt.Printf("%v\n", v.value)
		for otherV, _ := range v.succesors {
			fmt.Printf("  → %v\n", otherV.value)
		}
	}
}

const graphData string = `
nodes:
  "1": 123
  "2": 678
  "3": abc
  "4": xyz
arcs:
  "1":
    "2": 5
    "3": 1
  "2":
    "3": 9
  "3": {}
  "4":
    "2": 3
`
