package dfs

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/square"
)

func TestCarveIsATreeAndStopsAtTheOtherExit(t *testing.T) {
	for _, size := range []int{5, 30} {
		for seed := uint64(0); seed < 5; seed++ {
			m := mustSquare(t, size)
			start := m.NodesAt(0, 0)[0]
			other := m.NodesAt(size-1, size/2)[0]
			if err := m.SetExitPairs([]maze.ExitPair{{A: start, B: other}}); err != nil {
				t.Fatal(err)
			}
			if err := New(rand.New(rand.NewPCG(seed, uint64(size)))).Carve(m); err != nil {
				t.Fatalf("size %d seed %d: %v", size, seed, err)
			}
			if len(m.Nodes) != size*size {
				t.Fatalf("size %d seed %d: %d nodes", size, seed, len(m.Nodes))
			}
			if got := connectionCount(m); got != size*size-1 {
				t.Fatalf("size %d seed %d: %d connections, want %d", size, seed, got, size*size-1)
			}
			if len(other.Connections) != 1 {
				t.Fatalf("size %d seed %d: other exit has %d connections", size, seed, len(other.Connections))
			}
			if reached := walkConnections(start); reached != size*size {
				t.Fatalf("size %d seed %d: reached %d tiles from the first exit", size, seed, reached)
			}
		}
	}
}

func TestCarveStartsAtTheFirstExit(t *testing.T) {
	m := mustSquare(t, 8)
	start := m.NodesAt(0, 0)[0]
	other := m.NodesAt(0, 4)[0]
	if err := m.SetExitPairs([]maze.ExitPair{{A: start, B: other}}); err != nil {
		t.Fatal(err)
	}
	rec := &recorder{}
	if err := (carver{r: rec}).Carve(m); err != nil {
		t.Fatal(err)
	}
	if len(rec.calls) == 0 || rec.calls[0] != len(start.Walls)+len(start.Connections) {
		t.Fatalf("first choice was among %v, start now has %d open and %d closed edges", rec.calls, len(start.Connections), len(start.Walls))
	}
	if start.ConnTo(m.NodesAt(1, 0)[0]) == nil {
		t.Fatal("first step did not open the first neighbour of the first exit")
	}
	if len(other.Connections) != 1 {
		t.Fatalf("other exit has %d connections", len(other.Connections))
	}
}

func TestCarveRejectsTheWrongNumberOfPairs(t *testing.T) {
	m := mustSquare(t, 5)
	if err := New(rand.New(rand.NewPCG(1, 1))).Carve(m); err == nil {
		t.Fatal("expected an error with no exits")
	}
	a := m.NodesAt(0, 0)[0]
	b := m.NodesAt(0, 4)[0]
	c := m.NodesAt(4, 0)[0]
	d := m.NodesAt(4, 4)[0]
	if err := m.SetExitPairs([]maze.ExitPair{{A: a, B: b}, {A: c, B: d}}); err != nil {
		t.Fatal(err)
	}
	if err := New(rand.New(rand.NewPCG(1, 1))).Carve(m); err == nil {
		t.Fatal("expected an error with two exit pairs")
	}
}

// recorder always takes the first candidate and remembers how many there were.
type recorder struct {
	calls []int
}

func (r *recorder) IntN(n int) int {
	r.calls = append(r.calls, n)
	return 0
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

func connectionCount(m *maze.Maze) int {
	n := 0
	for _, node := range m.Nodes {
		n += len(node.Connections)
	}
	return n / 2
}

func walkConnections(start *maze.Node) int {
	seen := map[*maze.Node]bool{start: true}
	queue := []*maze.Node{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, e := range cur.Connections {
			next := e.Other(cur)
			if seen[next] {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return len(seen)
}
