package dot

import (
	"github.com/akualab/graph"
	graphviz "github.com/awalterschulze/gographviz"
	"testing"
)

// Reads a DOT file.
func TestDOTGraphREAD(t *testing.T) {

	parsed, err := graphviz.Parse([]byte(`
		digraph G {
			x -> 2 [ label = 5.1 ];
			4 -> 2 [ label = 1 ];
			4 -> x [ label = 2 ];
			x -> x [ label = 0.3 ];
		}

	`))
	if err != nil {
		panic(err)
	}

	dot := NewGraphDOT()
	graphviz.Analyse(parsed, dot)

	t.Logf("\n%v\n", dot.graph)
}

func TestConvertToDOT(t *testing.T) {

	g := sampleGraph(t)
	t.Logf("\n%s\n", DOT(g, "testing"))
}

func sampleGraph(t *testing.T) *graph.Graph {

	g := graph.New()

	// set some nodes
	g.Set("1", 123)
	g.Set("2", 678)
	g.Set("3", "abc")
	g.Set("4", "xyz")
	g.Set("xxx", "yyy")

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

	ok = g.Connect("4", "xxx", 1.11)
	if !ok {
		t.Fail()
	}

	return g
}
