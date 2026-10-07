// Package square is the first Topology: a square grid of walled tiles.
package square

import (
	"fmt"

	"github.com/jteutenberg/maze-gen/maze"
)

const (
	minSize = 5
	maxSize = 30
)

type topology struct {
	size int
}

// New returns a Topology that builds a size-by-size square.
// size must be between 5 and 30 inclusive.
func New(size int) (maze.Topology, error) {
	if size < minSize || size > maxSize {
		return nil, fmt.Errorf("size %d is outside %d..%d", size, minSize, maxSize)
	}
	return topology{size: size}, nil
}

func (t topology) Build() (*maze.Maze, error) {
	m := maze.NewMaze(t.size)
	nodes := make([]*maze.Node, t.size*t.size)
	for y := 0; y < t.size; y++ {
		for x := 0; x < t.size; x++ {
			border := x == 0 || y == 0 || x == t.size-1 || y == t.size-1
			n := m.AddNode(x, y, border)
			nodes[y*t.size+x] = n
		}
	}
	for y := 0; y < t.size; y++ {
		for x := 0; x < t.size-1; x++ {
			if _, err := m.AddWall(nodes[y*t.size+x], nodes[y*t.size+x+1]); err != nil {
				return nil, err
			}
		}
	}
	for y := 0; y < t.size-1; y++ {
		for x := 0; x < t.size; x++ {
			if _, err := m.AddWall(nodes[y*t.size+x], nodes[(y+1)*t.size+x]); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}
