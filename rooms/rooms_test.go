package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/square"
)

func TestMaxRooms(t *testing.T) {
	adder := New(rand.New(rand.NewPCG(1, 1)))
	cases := []struct{ size, want int }{{5, 0}, {6, 0}, {7, 1}, {12, 2}, {30, 5}}
	for _, c := range cases {
		m := mustSquare(t, c.size)
		if got := adder.MaxRooms(m); got != c.want {
			t.Errorf("size %d: MaxRooms = %d, want %d", c.size, got, c.want)
		}
	}
}

func TestAddRoomsOpensDisjointBlocks(t *testing.T) {
	m := mustSquare(t, 30)
	before := wallCount(m)
	adder := New(rand.New(rand.NewPCG(3, 4)))
	if err := adder.AddRooms(m, 5); err != nil {
		t.Fatal(err)
	}
	if got := wallCount(m); got != before-20 {
		t.Fatalf("walls = %d, want %d", got, before-20)
	}
	inRoom := 0
	for _, n := range m.Nodes {
		switch len(n.Connections) {
		case 0:
		case 2:
			inRoom++
		default:
			t.Fatalf("(%d,%d) has %d connections", n.X, n.Y, len(n.Connections))
		}
	}
	if inRoom != 20 {
		t.Fatalf("nodes in rooms = %d, want 20", inRoom)
	}
}

func TestRoomMayCoverBorderAndExit(t *testing.T) {
	m := mustSquare(t, 7)
	corner := m.NodesAt(0, 0)[0]
	far := m.NodesAt(6, 6)[0]
	if err := m.SetExitPairs([]maze.ExitPair{{A: corner, B: far}}); err != nil {
		t.Fatal(err)
	}
	nodes, ok := block(m, 0, 0)
	if !ok {
		t.Fatal("missing block")
	}
	if err := openBlock(m, nodes); err != nil {
		t.Fatal(err)
	}
	if len(corner.Connections) != 2 || !corner.Border {
		t.Fatalf("corner connections = %d, border = %v", len(corner.Connections), corner.Border)
	}
	if pairs := m.ExitPairs(); len(pairs) != 1 || pairs[0].A != corner {
		t.Fatal("opening a room dropped the exit")
	}
}

func TestAddRoomsRejectsOutOfRange(t *testing.T) {
	m := mustSquare(t, 7)
	adder := New(rand.New(rand.NewPCG(1, 1)))
	if err := adder.AddRooms(m, 2); err == nil {
		t.Fatal("expected error")
	}
	if err := adder.AddRooms(m, -1); err == nil {
		t.Fatal("expected error")
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

func wallCount(m *maze.Maze) int {
	n := 0
	for _, node := range m.Nodes {
		n += len(node.Walls)
	}
	return n / 2
}
