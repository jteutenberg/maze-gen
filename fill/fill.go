// Package fill carves the regions left after solutions and loops.
package fill

import (
	"fmt"
	"math/rand/v2"

	"github.com/jteutenberg/maze-gen/maze"
)

type filler struct {
	r *rand.Rand
}

// New returns a Filler. Each region is entered from one adjacent solution
// node, then carved with a randomised depth-first search. Branches stay open.
func New(r *rand.Rand) maze.Filler {
	return filler{r: r}
}

func (f filler) Fill(m *maze.Maze) error {
	for _, reg := range regions(m) {
		set := map[*maze.Node]bool{}
		for _, n := range reg {
			set[n] = true
		}
		start := f.entry(m, reg)
		visited := map[*maze.Node]bool{}
		if start == nil {
			start = reg[0]
			visited[start] = true
		}
		f.carve(m, start, set, visited)
		if len(visited) != len(reg) {
			return fmt.Errorf("filled %d of %d nodes in a region", len(visited), len(reg))
		}
	}
	return nil
}

func (f filler) entry(m *maze.Maze, reg []*maze.Node) *maze.Node {
	var far, near []*maze.Node
	seen := map[*maze.Node]bool{}
	for _, n := range reg {
		for _, nb := range neighbours(n) {
			if !nb.OnSolution || seen[nb] {
				continue
			}
			seen[nb] = true
			if m.FarFromExits(nb) {
				far = append(far, nb)
			} else {
				near = append(near, nb)
			}
		}
	}
	use := far
	if len(use) == 0 {
		use = near
	}
	if len(use) == 0 {
		return nil
	}
	return use[f.r.IntN(len(use))]
}

type hop struct {
	to   *maze.Node
	edge *maze.Edge
	wall bool
}

func (f filler) carve(m *maze.Maze, n *maze.Node, region, visited map[*maze.Node]bool) {
	var hops []hop
	for _, e := range n.Walls {
		to := e.Other(n)
		if region[to] && !visited[to] {
			hops = append(hops, hop{to: to, edge: e, wall: true})
		}
	}
	for _, e := range n.Connections {
		to := e.Other(n)
		if region[to] && !visited[to] {
			hops = append(hops, hop{to: to, edge: e, wall: false})
		}
	}
	f.r.Shuffle(len(hops), func(i, j int) {
		hops[i], hops[j] = hops[j], hops[i]
	})
	for _, h := range hops {
		if visited[h.to] {
			continue
		}
		if h.wall {
			_ = m.Open(h.edge)
		}
		visited[h.to] = true
		f.carve(m, h.to, region, visited)
	}
}

func regions(m *maze.Maze) [][]*maze.Node {
	seen := map[*maze.Node]bool{}
	var out [][]*maze.Node
	for _, n := range m.Nodes {
		if n.OnSolution || seen[n] {
			continue
		}
		seen[n] = true
		reg := []*maze.Node{n}
		for i := 0; i < len(reg); i++ {
			for _, nb := range neighbours(reg[i]) {
				if seen[nb] || nb.OnSolution {
					continue
				}
				seen[nb] = true
				reg = append(reg, nb)
			}
		}
		out = append(out, reg)
	}
	return out
}

func neighbours(n *maze.Node) []*maze.Node {
	out := make([]*maze.Node, 0, len(n.Walls)+len(n.Connections))
	for _, e := range n.Walls {
		out = append(out, e.Other(n))
	}
	for _, e := range n.Connections {
		out = append(out, e.Other(n))
	}
	return out
}
