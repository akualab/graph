// +build !goci

// Original work: Copyright (c) 2013 Alexander Willing, All rights reserved.
// Modified work: Copyright (c) 2013 AKUALAB INC., All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"launchpad.net/goyaml"
)

func TestYAML2(t *testing.T) {

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
		t.Fatal(ee)
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
		t.Fatal(eb)
	}
	fn = os.TempDir() + "graph2.yaml"
	err = ioutil.WriteFile(fn, b, 0644)
	if err != nil {
		panic(err)
	}

}

func TestJSON2(t *testing.T) {

	// Get the sample graph.
	g0 := sampleGraph(t)
	fn := os.TempDir() + "graph.json"
	t.Logf("json file: %s", fn)

	// Write YAML file.
	b, eb := json.Marshal(g0)
	if eb != nil {
		t.Fatal(eb)
	}
	err := ioutil.WriteFile(fn, b, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Read JSON file back and compare.
	dat, ee := ioutil.ReadFile(fn)
	if ee != nil {
		t.Fatal(ee)
	}

	g2 := New()
	err = json.Unmarshal(dat, g2)
	if err != nil {
		t.Fatal(err)
	}

	if e := compareGraphs(g0, g2); e != nil {
		t.Fatal(e)
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
