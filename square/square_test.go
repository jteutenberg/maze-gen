package square

import "testing"

func TestNewRejectsOutOfRange(t *testing.T) {
	for _, size := range []int{4, 31} {
		if _, err := New(size); err == nil {
			t.Errorf("size %d: expected error", size)
		}
	}
}

func TestBuildSquareOfFive(t *testing.T) {
	top, err := New(5)
	if err != nil {
		t.Fatal(err)
	}
	m, err := top.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Nodes) != 25 {
		t.Fatalf("nodes = %d, want 25", len(m.Nodes))
	}
	if got := len(m.BorderNodes()); got != 16 {
		t.Fatalf("border nodes = %d, want 16", got)
	}

	walls := 0
	for _, n := range m.Nodes {
		if len(n.Connections) != 0 {
			t.Fatalf("node (%d,%d) has a connection", n.X, n.Y)
		}
		want := 4
		if n.Border {
			want = 3
			if (n.X == 0 || n.X == 4) && (n.Y == 0 || n.Y == 4) {
				want = 2
			}
		}
		if len(n.Walls) != want {
			t.Errorf("(%d,%d) walls = %d, want %d", n.X, n.Y, len(n.Walls), want)
		}
		walls += len(n.Walls)
		for _, e := range n.Walls {
			other := e.Other(n)
			if other.WallTo(n) != e {
				t.Errorf("wall between (%d,%d) and (%d,%d) is not shared", n.X, n.Y, other.X, other.Y)
			}
		}
	}
	// Each undirected wall is stored on both endpoints.
	if walls != 80 {
		t.Fatalf("wall endpoints = %d, want 80", walls)
	}
	if got := m.NodesAt(0, 0); len(got) != 1 || got[0].X != 0 || got[0].Y != 0 {
		t.Fatalf("NodesAt(0,0) = %v", got)
	}
}
