// Package dfs carves passages by a depth-first search from the first exit.
package dfs

import (
	"fmt"
	"math/rand/v2"

	"github.com/jteutenberg/maze-gen/maze"
)

// Carver opens a passage from the first exit of the maze's only exit pair.
type Carver interface {
	Carve(m *maze.Maze) error
}

type source interface {
	IntN(n int) int
}

type carver struct {
	r source
}

// New returns a carver. At each step it chooses uniformly among the
// unvisited tiles that share a wall with the current tile, and opens that
// wall. The passage stays open. The search backtracks when the current tile
// has no unvisited neighbour, and when the current tile is the other exit.
// A visited tile is never entered again. No bridge is added.
func New(r *rand.Rand) Carver {
	return carver{r: r}
}

func (c carver) Carve(m *maze.Maze) error {
	pairs := m.ExitPairs()
	if len(pairs) != 1 {
		return fmt.Errorf("dfs needs one exit pair")
	}
	start, other := pairs[0].A, pairs[0].B
	if start == nil || other == nil {
		return fmt.Errorf("nil exit")
	}
	visited := map[*maze.Node]bool{}
	if err := c.walk(m, start, other, visited); err != nil {
		return err
	}
	if !visited[other] {
		return fmt.Errorf("dfs did not reach exit (%d,%d)", other.X, other.Y)
	}
	return nil
}

func (c carver) walk(m *maze.Maze, cur, other *maze.Node, visited map[*maze.Node]bool) error {
	visited[cur] = true
	if cur == other {
		return nil
	}
	for {
		candidates := unvisitedAdjacent(cur, visited)
		if len(candidates) == 0 {
			return nil
		}
		next := candidates[c.r.IntN(len(candidates))]
		edge := cur.WallTo(next)
		if edge == nil {
			return fmt.Errorf("no wall between (%d,%d) and (%d,%d)", cur.X, cur.Y, next.X, next.Y)
		}
		if err := m.Open(edge); err != nil {
			return err
		}
		if err := c.walk(m, next, other, visited); err != nil {
			return err
		}
	}
}

func unvisitedAdjacent(n *maze.Node, visited map[*maze.Node]bool) []*maze.Node {
	var out []*maze.Node
	for _, e := range n.Walls {
		other := e.Other(n)
		if !visited[other] {
			out = append(out, other)
		}
	}
	return out
}
