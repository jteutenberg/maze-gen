package maze

import "testing"

func TestSetExitPairsRejectsAdjacent(t *testing.T) {
	m := NewMaze(5)
	a := m.AddNode(0, 0, true)
	b := m.AddNode(1, 0, true)
	if _, err := m.AddWall(a, b); err != nil {
		t.Fatal(err)
	}
	err := m.SetExitPairs([]ExitPair{{A: a, B: b}})
	if err == nil {
		t.Fatal("expected adjacent exits to be rejected")
	}
}

func TestSetExitPairsAllowsDiagonal(t *testing.T) {
	m := NewMaze(5)
	a := m.AddNode(0, 0, true)
	b := m.AddNode(1, 1, false)
	if err := m.SetExitPairs([]ExitPair{{A: a, B: b}}); err != nil {
		t.Fatal(err)
	}
	if got := m.ExitPairs(); len(got) != 1 || got[0].A != a || got[0].B != b {
		t.Fatalf("pairs = %+v", got)
	}
}
