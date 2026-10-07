// Package loops carves extra solutions through regions that are not yet on a path.
package loops

import (
	"math/rand/v2"
	"sort"

	"github.com/jteutenberg/maze-gen/maze"
)

type adder struct {
	r *rand.Rand
}

// New returns a LoopAdder. A region is flooded through walls and connections,
// stopping at solution nodes.
func New(r *rand.Rand) maze.LoopAdder {
	return adder{r: r}
}

func (a adder) AddLoops(m *maze.Maze) error {
	queue := floodAll(m)
	for len(queue) > 0 {
		reg := queue[0]
		queue = queue[1:]
		path, anchors, ok := a.carveLoop(m, reg)
		if !ok {
			continue
		}
		for i, n := range path.nodes {
			n.OnSolution = true
			n.Path = path.id
			n.Index = i
		}
		if err := openAnchor(m, anchors[0].edge); err != nil {
			return err
		}
		if err := openAnchor(m, anchors[1].edge); err != nil {
			return err
		}
		m.AddSolution(maze.Solution{
			ID:        path.id,
			Nodes:     path.nodes,
			Waypoints: [2]*maze.Node{path.waypoint, nil},
		})
		queue = append(queue, remnants(reg)...)
	}
	return nil
}

// regions returns each connected set of non-solution nodes.
func regions(m *maze.Maze) [][]*maze.Node {
	return floodAll(m)
}

func floodAll(m *maze.Maze) [][]*maze.Node {
	seen := map[*maze.Node]bool{}
	var out [][]*maze.Node
	for _, n := range m.Nodes {
		if n.OnSolution || seen[n] {
			continue
		}
		out = append(out, flood(n, seen, nil))
	}
	return out
}

// remnants splits whatever is left of a region after a loop was carved out of it.
func remnants(reg []*maze.Node) [][]*maze.Node {
	set := map[*maze.Node]bool{}
	for _, n := range reg {
		if !n.OnSolution {
			set[n] = true
		}
	}
	seen := map[*maze.Node]bool{}
	var out [][]*maze.Node
	for n := range set {
		if seen[n] {
			continue
		}
		out = append(out, flood(n, seen, set))
	}
	return out
}

// flood walks from start through walls and connections. seen is updated.
// When limit is non-nil, only nodes in limit are entered.
func flood(start *maze.Node, seen map[*maze.Node]bool, limit map[*maze.Node]bool) []*maze.Node {
	seen[start] = true
	out := []*maze.Node{start}
	for i := 0; i < len(out); i++ {
		for _, nb := range neighbours(out[i]) {
			if seen[nb] || nb.OnSolution {
				continue
			}
			if limit != nil && !limit[nb] {
				continue
			}
			seen[nb] = true
			out = append(out, nb)
		}
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

type anchor struct {
	from *maze.Node
	sol  *maze.Node
	edge *maze.Edge
}

type carved struct {
	id       maze.PathID
	nodes    []*maze.Node
	waypoint *maze.Node
}

func (a adder) carveLoop(m *maze.Maze, reg []*maze.Node) (carved, [2]anchor, bool) {
	pairs := anchorPairs(reg)
	for i := range pairs {
		pairs[i].clear = m.FarFromExits(pairs[i].a.sol) && m.FarFromExits(pairs[i].b.sol)
	}
	a.r.Shuffle(len(pairs), func(i, j int) {
		pairs[i], pairs[j] = pairs[j], pairs[i]
	})
	// Prefer attachments at least two steps from an exit. A nearer pair is
	// used only when no farther pair can be carved. Larger index gaps still
	// come before smaller ones inside each group.
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].clear != pairs[j].clear {
			return pairs[i].clear
		}
		return pairs[i].diff > pairs[j].diff
	})
	allowed := map[*maze.Node]bool{}
	for _, n := range reg {
		allowed[n] = true
	}
	carves := 0
	for _, p := range pairs {
		// Pairs are sorted with index gaps of 3 or more first. A smaller gap
		// is tried only after those fail to produce a waypoint and a path.
		wpts := waypoints(reg, p.a.from, p.b.from, p.diff)
		if len(wpts) == 0 {
			continue
		}
		a.r.Shuffle(len(wpts), func(i, j int) {
			wpts[i], wpts[j] = wpts[j], wpts[i]
		})
		if len(wpts) > 8 {
			wpts = wpts[:8]
		}
		for _, w := range wpts {
			carves++
			nodes, ok := a.buildPath(m, allowed, p.a.from, w, p.b.from, p.diff)
			if ok {
				return carved{
					id:       maze.PathIDFor(p.a.from, p.b.from),
					nodes:    nodes,
					waypoint: w,
				}, [2]anchor{p.a, p.b}, true
			}
			if carves >= 64 {
				return carved{}, [2]anchor{}, false
			}
		}
	}
	return carved{}, [2]anchor{}, false
}

type anchored struct {
	a, b  anchor
	diff  int
	clear bool
}

func anchorPairs(reg []*maze.Node) []anchored {
	var anchors []anchor
	for _, n := range reg {
		anchors = append(anchors, touches(n)...)
	}
	var pairs []anchored
	for i := 0; i < len(anchors); i++ {
		for j := i + 1; j < len(anchors); j++ {
			a, b := anchors[i], anchors[j]
			if a.from == b.from || a.sol == b.sol || a.sol.Path != b.sol.Path {
				continue
			}
			diff := a.sol.Index - b.sol.Index
			if diff < 0 {
				diff = -diff
			}
			if diff < 1 {
				continue
			}
			pairs = append(pairs, anchored{a: a, b: b, diff: diff})
		}
	}
	return pairs
}

func touches(n *maze.Node) []anchor {
	var out []anchor
	seen := map[*maze.Node]bool{}
	consider := func(e *maze.Edge) {
		sol := e.Other(n)
		if !sol.OnSolution || seen[sol] {
			return
		}
		seen[sol] = true
		out = append(out, anchor{from: n, sol: sol, edge: e})
	}
	for _, e := range n.Walls {
		consider(e)
	}
	for _, e := range n.Connections {
		consider(e)
	}
	return out
}

func waypoints(reg []*maze.Node, a, b *maze.Node, diff int) []*maze.Node {
	need := diff / 2
	var out []*maze.Node
	for _, n := range reg {
		if n == a || n == b {
			continue
		}
		d1, d2 := manhattan(n, a), manhattan(n, b)
		if d1 >= need && d2 >= need && d1+d2 > diff {
			out = append(out, n)
		}
	}
	return out
}

func openAnchor(m *maze.Maze, e *maze.Edge) error {
	if e.A.ConnTo(e.B) == e {
		return nil
	}
	return m.Open(e)
}

func manhattan(a, b *maze.Node) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
