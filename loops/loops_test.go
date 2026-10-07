package loops

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/exits"
	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/solution"
	"github.com/jteutenberg/maze-gen/square"
)

func TestRoomInteriorIsOneRegion(t *testing.T) {
	m := mustSquare(t, 7)
	nodes, _ := room(m, 2, 2)
	for _, n := range m.Nodes {
		n.OnSolution = true
	}
	for _, n := range nodes {
		n.OnSolution = false
	}
	regs := regions(m)
	if len(regs) != 1 || len(regs[0]) != 4 {
		t.Fatalf("regions = %d, first = %d; want one region of 4", len(regs), len(regs[0]))
	}
}

func TestSeparatedNodesAreSeparateRegions(t *testing.T) {
	m := mustSquare(t, 5)
	for _, n := range m.Nodes {
		n.OnSolution = true
	}
	m.NodesAt(0, 0)[0].OnSolution = false
	m.NodesAt(2, 0)[0].OnSolution = false
	regs := regions(m)
	if len(regs) != 2 {
		t.Fatalf("regions = %d, want 2", len(regs))
	}
}

func TestLoopMeetsTheSolution(t *testing.T) {
	m := mustSquare(t, 10)
	markRow(m)
	ends := []maze.ExitPair{{A: m.NodesAt(0, 0)[0], B: m.NodesAt(m.Size-1, 0)[0]}}
	if err := m.SetExitPairs(ends); err != nil {
		t.Fatal(err)
	}
	before := len(m.Solutions())
	if err := New(rand.New(rand.NewPCG(1, 2))).AddLoops(m); err != nil {
		t.Fatal(err)
	}
	first := m.Solutions()[before]
	for _, nb := range attachments(first) {
		if !m.FarFromExits(nb) {
			t.Fatalf("loop attached at (%d,%d) index %d, next to an exit", nb.X, nb.Y, nb.Index)
		}
	}
	if len(m.Solutions()) <= before {
		t.Fatal("expected a loop")
	}
	assertSolutions(t, m)
	for _, loop := range m.Solutions()[before:] {
		diff, ok := attachmentDiff(loop)
		if !ok {
			a := loop.Nodes[0]
			b := loop.Nodes[len(loop.Nodes)-1]
			t.Fatalf("loop ends (%d,%d) and (%d,%d) do not meet one solution", a.X, a.Y, b.X, b.Y)
		}
		if len(loop.Nodes)-1 <= diff {
			t.Fatalf("loop edges = %d, attachment diff = %d", len(loop.Nodes)-1, diff)
		}
	}
}

func TestLoopsAfterGeneratedSolution(t *testing.T) {
	m := mustSquare(t, 12)
	r := rand.New(rand.NewPCG(4, 5))
	if err := exits.NewBorder(r).Assign(m); err != nil {
		t.Fatal(err)
	}
	if err := solution.New(r).Build(m); err != nil {
		t.Fatal(err)
	}
	if err := New(r).AddLoops(m); err != nil {
		t.Fatal(err)
	}
	assertSolutions(t, m)
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

func attachmentDiff(loop maze.Solution) (int, bool) {
	endA := attachmentsAt(loop.Nodes[0], loop.ID)
	endB := attachmentsAt(loop.Nodes[len(loop.Nodes)-1], loop.ID)
	for _, a := range endA {
		for _, b := range endB {
			if a.Path != b.Path || a == b {
				continue
			}
			diff := a.Index - b.Index
			if diff < 0 {
				diff = -diff
			}
			return diff, true
		}
	}
	return 0, false
}

func attachments(loop maze.Solution) []*maze.Node {
	return append(attachmentsAt(loop.Nodes[0], loop.ID), attachmentsAt(loop.Nodes[len(loop.Nodes)-1], loop.ID)...)
}

func attachmentsAt(n *maze.Node, self maze.PathID) []*maze.Node {
	var out []*maze.Node
	for _, e := range n.Connections {
		nb := e.Other(n)
		if nb.OnSolution && nb.Path != self {
			out = append(out, nb)
		}
	}
	return out
}

func assertSolutions(t *testing.T, m *maze.Maze) {
	t.Helper()
	seen := map[*maze.Node]bool{}
	for _, s := range m.Solutions() {
		if len(s.Nodes) < 2 {
			t.Fatal("short solution")
		}
		for j, n := range s.Nodes {
			if seen[n] {
				t.Fatalf("(%d,%d) is on two solutions", n.X, n.Y)
			}
			seen[n] = true
			if !n.OnSolution || n.Path != s.ID || n.Index != j {
				t.Fatalf("(%d,%d) index=%d path mismatch", n.X, n.Y, n.Index)
			}
			if j > 0 && n.ConnTo(s.Nodes[j-1]) == nil {
				t.Fatalf("gap before (%d,%d)", n.X, n.Y)
			}
		}
	}
}

func room(m *maze.Maze, x, y int) ([4]*maze.Node, bool) {
	coords := [4][2]int{{x, y}, {x + 1, y}, {x + 1, y + 1}, {x, y + 1}}
	var nodes [4]*maze.Node
	for i, c := range coords {
		at := m.NodesAt(c[0], c[1])
		if len(at) == 0 {
			return nodes, false
		}
		nodes[i] = at[0]
	}
	for i, n := range nodes {
		next := nodes[(i+1)%4]
		if err := m.Open(n.WallTo(next)); err != nil {
			return nodes, false
		}
	}
	return nodes, true
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
