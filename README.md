# Graph Data Structure in Go <a href="http://goci.me/project/github.com/akualab/graph"><img src="http://goci.me/project/image/github.com/akualab/graph" /></a>

**This project is a fork of https://github.com/sauerbraten/graph-store adapted to support directed graphs with weighted arcs. Removed locks.**

I removed gorouting synchronization because implementing and maintaining a goroutine-safe library is too complex. Concurrent access with writes can be synchronized externally. Mutliple goroutines reading is OK. This is the same approach used in the standard library.

Features:
* Directed graph with weighted arcs.
* Graph manipulation methods.
* A-Star search.
* IO support for GOB/JSON/YAML
* Arc weight normalization.

Coming soon:
* More graph manipulation methods (eg. merge graphs.)
* Viterbi search.

For more info about graphs visit https://en.wikipedia.org/wiki/Graph_(abstract_data_type)

## Usage

Get the package:

	$ go get github.com/akualab/graph

## Documentation

For full package documentation, visit http://godoc.org/github.com/akualab/graph


## License

BSD-like. See LICENSE file for details.
