// Package exits provides the step-2 strategies for choosing exit pairs.
package exits

import (
	"fmt"
	"math/rand/v2"

	"github.com/jteutenberg/maze-gen/maze"
)

const mixedAttempts = 10000

// NewBorder places one exit pair on two border nodes.
// The two exits may lie on the same side, including corners, so long as
// they do not share a wall.
func NewBorder(r *rand.Rand) maze.ExitAssigner {
	return pairAssigner{r: r, aOK: border, bOK: border, samePool: true}
}

// NewInterior places one exit pair on two interior nodes.
func NewInterior(r *rand.Rand) maze.ExitAssigner {
	return pairAssigner{r: r, aOK: interior, bOK: interior, samePool: true}
}

// NewMixed places two exit pairs: a border node with an interior node, and
// another interior node with another border node.
func NewMixed(r *rand.Rand) maze.ExitAssigner {
	return mixedAssigner{r: r}
}

// NewFixed places exits on the given positions and does not choose any.
// border and interior are each nil, or exactly two positions. Nil disables
// that kind of exit. When only one is set, those two nodes are the only
// pair. When both are set, the pairs have the same shape as NewMixed:
// border[0] with interior[0], and interior[1] with border[1].
func NewFixed(border, interior []maze.Pos) maze.ExitAssigner {
	return fixedAssigner{border: border, interior: interior}
}

type pairAssigner struct {
	r        *rand.Rand
	aOK      func(*maze.Node) bool
	bOK      func(*maze.Node) bool
	samePool bool
}

func (a pairAssigner) Assign(m *maze.Maze) error {
	pair, err := choosePair(a.r, m.Nodes, a.aOK, a.bOK, a.samePool)
	if err != nil {
		return err
	}
	return m.SetExitPairs([]maze.ExitPair{pair})
}

type fixedAssigner struct {
	border   []maze.Pos
	interior []maze.Pos
}

func (a fixedAssigner) Assign(m *maze.Maze) error {
	border, err := fixedNodes(m, a.border, true)
	if err != nil {
		return err
	}
	interior, err := fixedNodes(m, a.interior, false)
	if err != nil {
		return err
	}
	switch {
	case border != nil && interior != nil:
		return m.SetExitPairs([]maze.ExitPair{
			{A: border[0], B: interior[0]},
			{A: interior[1], B: border[1]},
		})
	case border != nil:
		return m.SetExitPairs([]maze.ExitPair{{A: border[0], B: border[1]}})
	case interior != nil:
		return m.SetExitPairs([]maze.ExitPair{{A: interior[0], B: interior[1]}})
	default:
		return fmt.Errorf("no exits")
	}
}

func fixedNodes(m *maze.Maze, positions []maze.Pos, wantBorder bool) ([]*maze.Node, error) {
	if positions == nil {
		return nil, nil
	}
	if len(positions) != 2 {
		return nil, fmt.Errorf("exit positions need two coordinates")
	}
	nodes := make([]*maze.Node, 2)
	for i, p := range positions {
		at := m.NodesAt(p.X, p.Y)
		if len(at) == 0 {
			return nil, fmt.Errorf("exit (%d,%d) is not in the maze", p.X, p.Y)
		}
		n := at[0]
		if wantBorder && !n.Border {
			return nil, fmt.Errorf("exit (%d,%d) is not on the border", n.X, n.Y)
		}
		if !wantBorder && n.Border {
			return nil, fmt.Errorf("exit (%d,%d) is not interior", n.X, n.Y)
		}
		nodes[i] = n
	}
	return nodes, nil
}

type mixedAssigner struct {
	r *rand.Rand
}

func (a mixedAssigner) Assign(m *maze.Maze) error {
	borders, interiors := split(m.Nodes)
	if len(borders) < 2 || len(interiors) < 2 {
		return fmt.Errorf("maze has %d border and %d interior nodes, need at least 2 of each", len(borders), len(interiors))
	}
	for range mixedAttempts {
		b1 := borders[a.r.IntN(len(borders))]
		b2 := borders[a.r.IntN(len(borders))]
		i1 := interiors[a.r.IntN(len(interiors))]
		i2 := interiors[a.r.IntN(len(interiors))]
		group := []*maze.Node{b1, i1, i2, b2}
		if !distinctAndSeparated(group) {
			continue
		}
		return m.SetExitPairs([]maze.ExitPair{
			{A: b1, B: i1},
			{A: i2, B: b2},
		})
	}
	return fmt.Errorf("could not place two separated exit pairs")
}

func choosePair(r *rand.Rand, nodes []*maze.Node, aOK, bOK func(*maze.Node) bool, samePool bool) (maze.ExitPair, error) {
	var pairs []maze.ExitPair
	for i, a := range nodes {
		if !aOK(a) {
			continue
		}
		for j, b := range nodes {
			if i == j || !bOK(b) || maze.Adjacent(a, b) {
				continue
			}
			if samePool && j < i {
				continue
			}
			pairs = append(pairs, maze.ExitPair{A: a, B: b})
		}
	}
	if len(pairs) == 0 {
		return maze.ExitPair{}, fmt.Errorf("no valid exit pair")
	}
	return pairs[r.IntN(len(pairs))], nil
}

func split(nodes []*maze.Node) (borders, interiors []*maze.Node) {
	for _, n := range nodes {
		if n.Border {
			borders = append(borders, n)
		} else {
			interiors = append(interiors, n)
		}
	}
	return borders, interiors
}

func distinctAndSeparated(nodes []*maze.Node) bool {
	for i, a := range nodes {
		for _, b := range nodes[i+1:] {
			if a == b || maze.Adjacent(a, b) {
				return false
			}
		}
	}
	return true
}

func border(n *maze.Node) bool   { return n.Border }
func interior(n *maze.Node) bool { return !n.Border }
