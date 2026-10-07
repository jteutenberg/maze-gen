package maze

// PathID is the identity of one solution: the positions of its two exits.
// The exit with the smaller x, then the smaller y, is stored first.
type PathID struct {
	AX, AY int
	BX, BY int
}

// Solution is the ordered path that joins one exit pair.
// Nodes[0] is exit A of the pair and Nodes[len-1] is exit B.
// Index on each node matches its position in Nodes.
type Solution struct {
	ID        PathID
	Nodes     []*Node
	Waypoints [2]*Node
}

// SolutionBuilder is step 4. It joins every exit pair with one solution.
// The command-line app does not ask a question for this step.
type SolutionBuilder interface {
	Build(m *Maze) error
}

// Solutions returns the solutions built by step 4, one per exit pair.
func (m *Maze) Solutions() []Solution {
	out := make([]Solution, len(m.solutions))
	copy(out, m.solutions)
	return out
}

// AddSolution appends a finished solution.
func (m *Maze) AddSolution(s Solution) {
	m.solutions = append(m.solutions, s)
}

// FarFromExits reports whether n is at least two steps along its solution
// from every exit on that solution. A node with no exit on its solution is far.
func (m *Maze) FarFromExits(n *Node) bool {
	if n == nil || !n.OnSolution {
		return true
	}
	exits := make(map[*Node]bool, len(m.exitPairs)*2)
	for _, p := range m.exitPairs {
		exits[p.A] = true
		exits[p.B] = true
	}
	var sol *Solution
	for i := range m.solutions {
		if m.solutions[i].ID == n.Path {
			sol = &m.solutions[i]
			break
		}
	}
	if sol == nil {
		return true
	}
	nearest := -1
	for _, e := range sol.Nodes {
		if !exits[e] {
			continue
		}
		d := n.Index - e.Index
		if d < 0 {
			d = -d
		}
		if nearest < 0 || d < nearest {
			nearest = d
		}
	}
	if nearest < 0 {
		return true
	}
	return nearest >= 2
}

// PathIDFor returns the canonical identity of the solution that joins a and b.
func PathIDFor(a, b *Node) PathID {
	if b.X < a.X || (b.X == a.X && b.Y < a.Y) {
		a, b = b, a
	}
	return PathID{AX: a.X, AY: a.Y, BX: b.X, BY: b.Y}
}
