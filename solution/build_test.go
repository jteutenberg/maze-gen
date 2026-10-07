package solution

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/exits"
	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/rooms"
	"github.com/jteutenberg/maze-gen/square"
)

func TestSolutionJoinsBorderExits(t *testing.T) {
	for _, size := range []int{5, 10, 30} {
		for seed := uint64(1); seed <= 3; seed++ {
			m := mustSquare(t, size)
			r := rand.New(rand.NewPCG(seed, uint64(size)))
			if err := exits.NewBorder(r).Assign(m); err != nil {
				t.Fatal(err)
			}
			before := connectionCount(m)
			if err := New(r).Build(m); err != nil {
				t.Fatalf("size %d seed %d: %v", size, seed, err)
			}
			assertSolutions(t, m, before)
		}
	}
}

func TestSolutionJoinsTwoPairsAndRooms(t *testing.T) {
	m := mustSquare(t, 15)
	r := rand.New(rand.NewPCG(9, 9))
	if err := exits.NewMixed(r).Assign(m); err != nil {
		t.Fatal(err)
	}
	adder := rooms.New(r)
	if err := adder.AddRooms(m, adder.MaxRooms(m)); err != nil {
		t.Fatal(err)
	}
	before := connectionCount(m)
	if err := New(r).Build(m); err != nil {
		t.Fatal(err)
	}
	assertSolutions(t, m, before)
}

func TestBuildFallsBackToOneWaypoint(t *testing.T) {
	m := mustSquare(t, 7)
	a := m.NodesAt(0, 0)[0]
	b := m.NodesAt(6, 0)[0]
	for _, n := range m.Nodes {
		if n.Y != 0 {
			n.OnSolution = true
		}
	}
	if err := m.SetExitPairs([]maze.ExitPair{{A: a, B: b}}); err != nil {
		t.Fatal(err)
	}
	if err := New(rand.New(rand.NewPCG(1, 1))).Build(m); err != nil {
		t.Fatal(err)
	}
	sol := m.Solutions()[0]
	if sol.Waypoints[0] != m.NodesAt(3, 0)[0] || sol.Waypoints[1] != nil {
		t.Fatalf("waypoints = (%v, %v)", sol.Waypoints[0], sol.Waypoints[1])
	}
	if sol.Nodes[0] != a || sol.Nodes[len(sol.Nodes)-1] != b {
		t.Fatal("path does not join the exits")
	}
}

func TestBuildFallsBackToDirect(t *testing.T) {
	m := mustSquare(t, 8)
	a := m.NodesAt(0, 0)[0]
	mid := m.NodesAt(1, 0)[0]
	b := m.NodesAt(2, 0)[0]
	for _, n := range m.Nodes {
		if n != a && n != mid && n != b {
			n.OnSolution = true
		}
	}
	if err := m.SetExitPairs([]maze.ExitPair{{A: a, B: b}}); err != nil {
		t.Fatal(err)
	}
	if err := New(rand.New(rand.NewPCG(1, 1))).Build(m); err != nil {
		t.Fatal(err)
	}
	sol := m.Solutions()[0]
	if sol.Waypoints[0] != nil || sol.Waypoints[1] != nil {
		t.Fatal("direct path kept a waypoint")
	}
	if sol.Nodes[0] != a || sol.Nodes[len(sol.Nodes)-1] != b {
		t.Fatal("path does not join the exits")
	}
}

func TestBuildReturnsErrorWhenNoRoute(t *testing.T) {
	m := mustSquare(t, 5)
	a := m.NodesAt(0, 0)[0]
	mid := m.NodesAt(1, 0)[0]
	b := m.NodesAt(2, 0)[0]
	for _, n := range m.Nodes {
		if n != a && n != b {
			n.OnSolution = true
		}
	}
	if err := m.RemoveWall(a.WallTo(mid)); err != nil {
		t.Fatal(err)
	}
	if err := m.RemoveWall(mid.WallTo(b)); err != nil {
		t.Fatal(err)
	}
	if err := m.SetExitPairs([]maze.ExitPair{{A: a, B: b}}); err != nil {
		t.Fatal(err)
	}
	if err := New(rand.New(rand.NewPCG(1, 1))).Build(m); err == nil {
		t.Fatal("expected no route between the exits")
	}
}

func TestBridgeUndoRestoresTheUnderPath(t *testing.T) {
	m := mustSquare(t, 5)
	cur := m.NodesAt(1, 2)[0]
	mid := m.NodesAt(2, 2)[0]
	far := m.NodesAt(3, 2)[0]
	id := maze.PathID{AX: 2, AY: 0, BX: 2, BY: 4}
	mark(mid, id)
	walls := wallCount(m)
	nodes := len(m.Nodes)
	st, err := applyBridge(m, cur, mid, far, maze.PathID{AX: 0, AY: 2, BX: 4, BY: 2}, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !far.OnSolution || !st.bridge.OnSolution {
		t.Fatal("bridge ends were not marked on the crossing solution")
	}
	if mid.Path != id {
		t.Fatal("bridge rewrote the under path")
	}
	if cur.ConnTo(mid) != nil || mid.ConnTo(far) != nil {
		t.Fatal("under path was connected across the bridge")
	}
	if cur.ConnTo(st.bridge) == nil || st.bridge.ConnTo(far) == nil {
		t.Fatal("bridge is not connected to both outer nodes")
	}
	if st.bridge.X != mid.X || st.bridge.Y != mid.Y {
		t.Fatalf("bridge position = (%d,%d)", st.bridge.X, st.bridge.Y)
	}
	undoStep(m, cur, st)
	if far.OnSolution || len(m.Nodes) != nodes || wallCount(m) != walls {
		t.Fatalf("undo left nodes=%d walls=%d farOn=%v", len(m.Nodes), wallCount(m), far.OnSolution)
	}
	if cur.WallTo(mid) == nil || mid.WallTo(far) == nil {
		t.Fatal("undo did not restore the crossed walls")
	}
}

func assertSolutions(t *testing.T, m *maze.Maze, connectionsBefore int) {
	t.Helper()
	pairs := m.ExitPairs()
	sols := m.Solutions()
	if len(sols) != len(pairs) {
		t.Fatalf("got %d solutions, want %d", len(sols), len(pairs))
	}
	seen := map[*maze.Node]bool{}
	added := 0
	minDist := m.Size / 2
	for i, s := range sols {
		pair := pairs[i]
		if len(s.Nodes) < 2 || s.Nodes[0] != pair.A || s.Nodes[len(s.Nodes)-1] != pair.B {
			t.Fatalf("solution %d does not run from exit A to exit B", i)
		}
		bridges := 0
		for j, n := range s.Nodes {
			if seen[n] {
				t.Fatalf("node (%d,%d) is on two solutions", n.X, n.Y)
			}
			seen[n] = true
			if !n.OnSolution || n.Path != s.ID || n.Index != j {
				t.Fatalf("node (%d,%d) index=%d on=%v", n.X, n.Y, n.Index, n.OnSolution)
			}
			if j > 0 && n.ConnTo(s.Nodes[j-1]) == nil {
				t.Fatalf("gap before (%d,%d)", n.X, n.Y)
			}
			at := m.NodesAt(n.X, n.Y)
			if len(at) > 0 && at[0] != n {
				bridges++
			}
		}
		if bridges > 2 {
			t.Fatalf("solution %d has %d bridges", i, bridges)
		}
		assertExitBend(t, s.Nodes)
		w1, w2 := s.Waypoints[0], s.Waypoints[1]
		if w1 != nil && (manhattan(w1, pair.A) < minDist || manhattan(w1, pair.B) < minDist) {
			t.Fatalf("waypoint (%d,%d) is too close to an exit", w1.X, w1.Y)
		}
		if w2 != nil && (manhattan(w2, pair.A) < minDist || manhattan(w2, pair.B) < minDist || (w1 != nil && manhattan(w1, w2) < minDist)) {
			t.Fatalf("waypoint (%d,%d) is too close", w2.X, w2.Y)
		}
		added += len(s.Nodes) - 1
	}
	got := connectionCount(m) - connectionsBefore
	if got > added || (connectionsBefore == 0 && got != added) {
		t.Fatalf("new connections = %d, path edges = %d", got, added)
	}
}

func assertExitBend(t *testing.T, nodes []*maze.Node) {
	t.Helper()
	if len(nodes) < 3 {
		return
	}
	if collinear(nodes[0], nodes[1], nodes[2]) {
		t.Fatalf("exit (%d,%d) is in line with (%d,%d) and (%d,%d)",
			nodes[0].X, nodes[0].Y, nodes[1].X, nodes[1].Y, nodes[2].X, nodes[2].Y)
	}
	n := len(nodes)
	if collinear(nodes[n-1], nodes[n-2], nodes[n-3]) {
		t.Fatalf("exit (%d,%d) is in line with (%d,%d) and (%d,%d)",
			nodes[n-1].X, nodes[n-1].Y, nodes[n-2].X, nodes[n-2].Y, nodes[n-3].X, nodes[n-3].Y)
	}
}

func collinear(a, b, c *maze.Node) bool {
	return (b.X-a.X)*(c.Y-a.Y) == (b.Y-a.Y)*(c.X-a.X)
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

func wallCount(m *maze.Maze) int {
	n := 0
	for _, node := range m.Nodes {
		n += len(node.Walls)
	}
	return n / 2
}
