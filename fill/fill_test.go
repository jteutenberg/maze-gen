package fill

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/exits"
	"github.com/jteutenberg/maze-gen/loops"
	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/solution"
	"github.com/jteutenberg/maze-gen/square"
)

func TestFillReachesEveryNode(t *testing.T) {
	for seed := uint64(1); seed <= 3; seed++ {
		m := mustSquare(t, 10)
		r := rand.New(rand.NewPCG(seed, 7))
		if err := exits.NewBorder(r).Assign(m); err != nil {
			t.Fatal(err)
		}
		if err := solution.New(r).Build(m); err != nil {
			t.Fatal(err)
		}
		if err := loops.New(r).AddLoops(m); err != nil {
			t.Fatal(err)
		}
		if err := New(r).Fill(m); err != nil {
			t.Fatal(err)
		}
		if got, want := reached(m), len(m.Nodes); got != want {
			t.Fatalf("seed %d: reached %d of %d nodes", seed, got, want)
		}
	}
}

func TestEntryStaysClearOfExits(t *testing.T) {
	m := mustSquare(t, 8)
	markRow(m)
	if err := m.SetExitPairs([]maze.ExitPair{{A: m.NodesAt(0, 0)[0], B: m.NodesAt(m.Size-1, 0)[0]}}); err != nil {
		t.Fatal(err)
	}
	reg := regions(m)[0]
	for seed := uint64(0); seed < 20; seed++ {
		start := filler{r: rand.New(rand.NewPCG(seed, 1))}.entry(m, reg)
		if start == nil || !m.FarFromExits(start) {
			t.Fatalf("seed %d: entry = %v", seed, start)
		}
	}
}

func TestEntryFallsBackBesideAnExit(t *testing.T) {
	m := mustSquare(t, 5)
	exit := m.NodesAt(0, 0)[0]
	other := m.NodesAt(4, 0)[0]
	exit.OnSolution = true
	other.OnSolution = true
	exit.Path = maze.PathID{BX: 4}
	other.Path = exit.Path
	exit.Index = 0
	other.Index = 4
	m.AddSolution(maze.Solution{ID: exit.Path, Nodes: []*maze.Node{exit, other}})
	if err := m.SetExitPairs([]maze.ExitPair{{A: exit, B: other}}); err != nil {
		t.Fatal(err)
	}
	cell := m.NodesAt(0, 1)[0]
	start := filler{r: rand.New(rand.NewPCG(1, 1))}.entry(m, []*maze.Node{cell})
	if start != exit {
		t.Fatalf("entry = (%d,%d), want the only door beside the exit", start.X, start.Y)
	}
}

func markRow(m *maze.Maze) {
	id := maze.PathID{BX: m.Size - 1}
	nodes := make([]*maze.Node, m.Size)
	for x := 0; x < m.Size; x++ {
		n := m.NodesAt(x, 0)[0]
		n.OnSolution = true
		n.Path = id
		n.Index = x
		nodes[x] = n
	}
	for x := 0; x < m.Size-1; x++ {
		_ = m.Open(nodes[x].WallTo(nodes[x+1]))
	}
	m.AddSolution(maze.Solution{ID: id, Nodes: nodes})
}

func reached(m *maze.Maze) int {
	seen := map[*maze.Node]bool{}
	var q []*maze.Node
	for _, n := range m.Nodes {
		if n.OnSolution {
			seen[n] = true
			q = append(q, n)
		}
	}
	for i := 0; i < len(q); i++ {
		for _, e := range q[i].Connections {
			nb := e.Other(q[i])
			if seen[nb] {
				continue
			}
			seen[nb] = true
			q = append(q, nb)
		}
	}
	return len(seen)
}

func mustSquare(t *testing.T, size int) *maze.Maze {
	t.Helper()
	top, err := square.New(size)
	if err != nil {
		t.Fatal(err)
	}
	m, err := top.Build()
	if err != nil {
		t.Fatal(err)
	}
	return m
}
