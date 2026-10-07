package maze

import "testing"

func TestFarFromExits(t *testing.T) {
	m := NewMaze(6)
	id := PathID{BX: 5}
	nodes := make([]*Node, 6)
	for x := 0; x < 6; x++ {
		n := m.AddNode(x, 0, true)
		n.OnSolution = true
		n.Path = id
		n.Index = x
		nodes[x] = n
	}
	m.AddSolution(Solution{ID: id, Nodes: nodes})
	if err := m.SetExitPairs([]ExitPair{{A: nodes[0], B: nodes[5]}}); err != nil {
		t.Fatal(err)
	}
	for _, i := range []int{0, 1, 4, 5} {
		if m.FarFromExits(nodes[i]) {
			t.Fatalf("index %d is in sight of an exit", i)
		}
	}
	for _, i := range []int{2, 3} {
		if !m.FarFromExits(nodes[i]) {
			t.Fatalf("index %d should be two steps from both exits", i)
		}
	}
}
