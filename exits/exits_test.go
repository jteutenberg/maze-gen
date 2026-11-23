package exits

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/square"
)

func TestBorderPair(t *testing.T) {
	for _, size := range []int{5, 30} {
		m := mustSquare(t, size)
		r := rand.New(rand.NewPCG(1, uint64(size)))
		if err := NewBorder(r).Assign(m); err != nil {
			t.Fatal(err)
		}
		assertPairs(t, m.ExitPairs(), 1, func(n *maze.Node) bool { return n.Border })
	}
}

func TestInteriorPair(t *testing.T) {
	for _, size := range []int{5, 30} {
		m := mustSquare(t, size)
		r := rand.New(rand.NewPCG(2, uint64(size)))
		if err := NewInterior(r).Assign(m); err != nil {
			t.Fatal(err)
		}
		assertPairs(t, m.ExitPairs(), 1, func(n *maze.Node) bool { return !n.Border })
	}
}

func TestFixedBorderPair(t *testing.T) {
	m := mustSquare(t, 8)
	err := NewFixed([]maze.Pos{{X: 0, Y: 0}, {X: 0, Y: 7}}, nil).Assign(m)
	if err != nil {
		t.Fatal(err)
	}
	pairs := m.ExitPairs()
	if len(pairs) != 1 || pairs[0].A != m.NodesAt(0, 0)[0] || pairs[0].B != m.NodesAt(0, 7)[0] {
		t.Fatalf("pairs = %v", pairCoords(pairs))
	}
}

func TestFixedInteriorPair(t *testing.T) {
	m := mustSquare(t, 8)
	err := NewFixed(nil, []maze.Pos{{X: 2, Y: 2}, {X: 5, Y: 5}}).Assign(m)
	if err != nil {
		t.Fatal(err)
	}
	pairs := m.ExitPairs()
	if len(pairs) != 1 || pairs[0].A != m.NodesAt(2, 2)[0] || pairs[0].B != m.NodesAt(5, 5)[0] {
		t.Fatalf("pairs = %v", pairCoords(pairs))
	}
}

func TestFixedMixedPairs(t *testing.T) {
	m := mustSquare(t, 8)
	err := NewFixed(
		[]maze.Pos{{X: 0, Y: 0}, {X: 7, Y: 7}},
		[]maze.Pos{{X: 2, Y: 2}, {X: 5, Y: 5}},
	).Assign(m)
	if err != nil {
		t.Fatal(err)
	}
	pairs := m.ExitPairs()
	if len(pairs) != 2 {
		t.Fatalf("pairs = %v", pairCoords(pairs))
	}
	if pairs[0].A != m.NodesAt(0, 0)[0] || pairs[0].B != m.NodesAt(2, 2)[0] {
		t.Fatalf("first pair = %v", pairCoords(pairs[:1]))
	}
	if pairs[1].A != m.NodesAt(5, 5)[0] || pairs[1].B != m.NodesAt(7, 7)[0] {
		t.Fatalf("second pair = %v", pairCoords(pairs[1:]))
	}
	assertSeparated(t, []*maze.Node{pairs[0].A, pairs[0].B, pairs[1].A, pairs[1].B})
}

func TestFixedRejectsWrongKind(t *testing.T) {
	m := mustSquare(t, 8)
	if err := NewFixed([]maze.Pos{{X: 1, Y: 1}, {X: 0, Y: 7}}, nil).Assign(m); err == nil {
		t.Fatal("interior node accepted as a border exit")
	}
	if err := NewFixed(nil, []maze.Pos{{X: 0, Y: 0}, {X: 2, Y: 2}}).Assign(m); err == nil {
		t.Fatal("border node accepted as an interior exit")
	}
}

func TestFixedRejectsMissingAdjacentAndRepeated(t *testing.T) {
	m := mustSquare(t, 8)
	if err := NewFixed([]maze.Pos{{X: 0, Y: 0}, {X: 9, Y: 0}}, nil).Assign(m); err == nil {
		t.Fatal("missing tile accepted")
	}
	if err := NewFixed([]maze.Pos{{X: 0, Y: 0}, {X: 1, Y: 0}}, nil).Assign(m); err == nil {
		t.Fatal("adjacent exits accepted")
	}
	if err := NewFixed([]maze.Pos{{X: 0, Y: 0}, {X: 0, Y: 0}}, nil).Assign(m); err == nil {
		t.Fatal("repeated exit accepted")
	}
	if err := NewFixed(
		[]maze.Pos{{X: 0, Y: 2}, {X: 7, Y: 7}},
		[]maze.Pos{{X: 1, Y: 2}, {X: 5, Y: 5}},
	).Assign(m); err == nil {
		t.Fatal("adjacent exits from different flags accepted")
	}
	if err := NewFixed(nil, nil).Assign(m); err == nil {
		t.Fatal("no exits accepted")
	}
	if err := NewFixed([]maze.Pos{{X: 0, Y: 0}}, nil).Assign(m); err == nil {
		t.Fatal("one coordinate accepted")
	}
}

func pairCoords(pairs []maze.ExitPair) [][4]int {
	out := make([][4]int, len(pairs))
	for i, p := range pairs {
		out[i] = [4]int{p.A.X, p.A.Y, p.B.X, p.B.Y}
	}
	return out
}

func TestMixedPairs(t *testing.T) {
	for _, size := range []int{5, 30} {
		for seed := uint64(0); seed < 20; seed++ {
			m := mustSquare(t, size)
			r := rand.New(rand.NewPCG(seed, uint64(size)))
			if err := NewMixed(r).Assign(m); err != nil {
				t.Fatalf("size %d seed %d: %v", size, seed, err)
			}
			pairs := m.ExitPairs()
			if len(pairs) != 2 {
				t.Fatalf("size %d seed %d: %d pairs", size, seed, len(pairs))
			}
			if !pairs[0].A.Border || pairs[0].B.Border || pairs[1].A.Border || !pairs[1].B.Border {
				t.Fatalf("size %d seed %d: roles = (%v,%v) (%v,%v)", size, seed,
					pairs[0].A.Border, pairs[0].B.Border, pairs[1].A.Border, pairs[1].B.Border)
			}
			assertSeparated(t, []*maze.Node{pairs[0].A, pairs[0].B, pairs[1].A, pairs[1].B})
		}
	}
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

func assertPairs(t *testing.T, pairs []maze.ExitPair, n int, ok func(*maze.Node) bool) {
	t.Helper()
	if len(pairs) != n {
		t.Fatalf("got %d pairs, want %d", len(pairs), n)
	}
	var nodes []*maze.Node
	for _, p := range pairs {
		if !ok(p.A) || !ok(p.B) {
			t.Fatalf("pair (%d,%d)-(%d,%d) has the wrong kind of node", p.A.X, p.A.Y, p.B.X, p.B.Y)
		}
		nodes = append(nodes, p.A, p.B)
	}
	assertSeparated(t, nodes)
}

func assertSeparated(t *testing.T, nodes []*maze.Node) {
	t.Helper()
	for i, a := range nodes {
		for _, b := range nodes[i+1:] {
			if a == b {
				t.Fatalf("exit (%d,%d) repeated", a.X, a.Y)
			}
			if maze.Adjacent(a, b) {
				t.Fatalf("exits (%d,%d) and (%d,%d) are adjacent", a.X, a.Y, b.X, b.Y)
			}
		}
	}
}
