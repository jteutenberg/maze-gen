package maze

import "fmt"

// ExitPair is two exits joined by one solution path.
// A is one endpoint and B is the other. Order within the pair does not
// make one of them the start.
type ExitPair struct {
	A, B *Node
}

// ExitAssigner is step 2. It chooses one or two exit pairs and stores them
// on the maze. The command-line app picks an implementation; Assign takes
// no further input.
type ExitAssigner interface {
	Assign(m *Maze) error
}

// ExitPairs returns the pairs stored by step 2.
func (m *Maze) ExitPairs() []ExitPair {
	out := make([]ExitPair, len(m.exitPairs))
	copy(out, m.exitPairs)
	return out
}

// SetExitPairs replaces the maze's exit pairs.
// Every exit must be a node of this maze, all exits must be distinct, and
// no two exits may be adjacent.
func (m *Maze) SetExitPairs(pairs []ExitPair) error {
	if len(pairs) == 0 {
		return fmt.Errorf("no exit pairs")
	}
	exits := make([]*Node, 0, len(pairs)*2)
	for _, p := range pairs {
		exits = append(exits, p.A, p.B)
	}
	seen := make(map[*Node]bool, len(exits))
	for _, n := range exits {
		if n == nil {
			return fmt.Errorf("nil exit")
		}
		if !m.contains(n) {
			return fmt.Errorf("exit (%d,%d) is not in the maze", n.X, n.Y)
		}
		if seen[n] {
			return fmt.Errorf("exit (%d,%d) is used more than once", n.X, n.Y)
		}
		seen[n] = true
	}
	for i, a := range exits {
		for _, b := range exits[i+1:] {
			if Adjacent(a, b) {
				return fmt.Errorf("exits (%d,%d) and (%d,%d) are adjacent", a.X, a.Y, b.X, b.Y)
			}
		}
	}
	m.exitPairs = append([]ExitPair(nil), pairs...)
	return nil
}

func (m *Maze) contains(n *Node) bool {
	for _, cand := range m.at[Pos{X: n.X, Y: n.Y}] {
		if cand == n {
			return true
		}
	}
	return false
}
