// Package rooms opens 2×2 blocks of walls into connections.
package rooms

import (
	"fmt"
	"math/rand/v2"

	"github.com/jteutenberg/maze-gen/maze"
)

type adder struct {
	r *rand.Rand
}

// New returns a RoomAdder that places up to (size-2)/5 rooms.
// Rooms share no nodes. They may lie on the border, contain an exit, and
// touch along an outer edge.
func New(r *rand.Rand) maze.RoomAdder {
	return adder{r: r}
}

func (adder) MaxRooms(m *maze.Maze) int {
	n := (m.Size - 2) / 5
	if n < 0 {
		return 0
	}
	return n
}

func (a adder) AddRooms(m *maze.Maze, n int) error {
	max := a.MaxRooms(m)
	if n < 0 || n > max {
		return fmt.Errorf("%d rooms is outside 0..%d", n, max)
	}
	if n == 0 {
		return nil
	}
	origins := make([][2]int, 0, (m.Size-1)*(m.Size-1))
	for y := 0; y <= m.Size-2; y++ {
		for x := 0; x <= m.Size-2; x++ {
			origins = append(origins, [2]int{x, y})
		}
	}
	a.r.Shuffle(len(origins), func(i, j int) {
		origins[i], origins[j] = origins[j], origins[i]
	})

	used := map[*maze.Node]bool{}
	placed := 0
	for _, o := range origins {
		nodes, ok := block(m, o[0], o[1])
		if !ok || overlaps(used, nodes) {
			continue
		}
		if err := openBlock(m, nodes); err != nil {
			return err
		}
		for _, n := range nodes {
			used[n] = true
		}
		placed++
		if placed == n {
			return nil
		}
	}
	return fmt.Errorf("placed %d of %d rooms", placed, n)
}

func block(m *maze.Maze, x, y int) ([4]*maze.Node, bool) {
	coords := [4][2]int{{x, y}, {x + 1, y}, {x + 1, y + 1}, {x, y + 1}}
	var nodes [4]*maze.Node
	for i, c := range coords {
		at := m.NodesAt(c[0], c[1])
		if len(at) == 0 {
			return nodes, false
		}
		nodes[i] = at[0]
	}
	return nodes, true
}

func overlaps(used map[*maze.Node]bool, nodes [4]*maze.Node) bool {
	for _, n := range nodes {
		if used[n] {
			return true
		}
	}
	return false
}

func openBlock(m *maze.Maze, nodes [4]*maze.Node) error {
	for i, n := range nodes {
		next := nodes[(i+1)%4]
		e := n.WallTo(next)
		if e == nil {
			return fmt.Errorf("no wall between (%d,%d) and (%d,%d)", n.X, n.Y, next.X, next.Y)
		}
		if err := m.Open(e); err != nil {
			return err
		}
	}
	return nil
}
