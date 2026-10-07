package loops

import (
	"errors"

	"github.com/jteutenberg/maze-gen/maze"
)

var errNoPath = errors.New("no path")

type leg struct {
	start       *maze.Node
	markedStart bool
	steps       []step
}

type step struct {
	node   *maze.Node
	edge   *maze.Edge
	opened bool
	marked bool
	dx, dy int
}

func (l leg) nodes() []*maze.Node {
	out := []*maze.Node{l.start}
	for _, s := range l.steps {
		out = append(out, s.node)
	}
	return out
}

func (l leg) undo(m *maze.Maze) {
	for i := len(l.steps) - 1; i >= 0; i-- {
		prev := l.start
		if i > 0 {
			prev = l.steps[i-1].node
		}
		undoStep(m, stPrev(prev, l.steps[i]))
	}
	if l.markedStart {
		unmark(l.start)
	}
}

func stPrev(prev *maze.Node, st step) undo {
	return undo{prev: prev, step: st}
}

type undo struct {
	prev *maze.Node
	step step
}

func undoStep(m *maze.Maze, u undo) {
	if u.step.opened {
		_ = m.Close(u.step.edge)
	}
	if u.step.marked {
		unmark(u.step.node)
	}
}

func (a adder) buildPath(m *maze.Maze, allowed map[*maze.Node]bool, start, w, dest *maze.Node, diff int) ([]*maze.Node, bool) {
	id := maze.PathIDFor(start, dest)
	first, err := a.walk(m, allowed, start, w, dest, id)
	if err != nil {
		return nil, false
	}
	second, err := a.walk(m, allowed, w, dest, start, id)
	if err != nil {
		first.undo(m)
		return nil, false
	}
	nodes := first.nodes()
	rest := second.nodes()
	nodes = append(nodes, rest[1:]...)
	if len(nodes)-1 <= diff {
		second.undo(m)
		first.undo(m)
		return nil, false
	}
	return nodes, true
}

func (a adder) walk(m *maze.Maze, allowed map[*maze.Node]bool, start, dest, forbid *maze.Node, id maze.PathID) (leg, error) {
	lg := leg{start: start}
	if !start.OnSolution {
		mark(start, id)
		lg.markedStart = true
	}
	search := map[*maze.Node]bool{start: true}
	limit := len(allowed)*8 + 64
	for guard := 0; guard < limit; guard++ {
		cur := start
		if len(lg.steps) > 0 {
			cur = lg.steps[len(lg.steps)-1].node
		}
		if cur == dest {
			return lg, nil
		}
		mv, ok := a.pick(cur, dest, forbid, search, allowed)
		if !ok {
			if len(lg.steps) == 0 {
				lg.undo(m)
				return leg{}, errNoPath
			}
			last := lg.steps[len(lg.steps)-1]
			undoStep(m, undo{prev: previous(start, lg.steps), step: last})
			lg.steps = lg.steps[:len(lg.steps)-1]
			continue
		}
		st := step{node: mv.to, edge: mv.edge, dx: mv.dx, dy: mv.dy}
		if mv.wall {
			if err := m.Open(mv.edge); err != nil {
				lg.undo(m)
				return leg{}, err
			}
			st.opened = true
		}
		if !mv.to.OnSolution {
			mark(mv.to, id)
			st.marked = true
		}
		search[mv.to] = true
		lg.steps = append(lg.steps, st)
	}
	lg.undo(m)
	return leg{}, errNoPath
}

func previous(start *maze.Node, steps []step) *maze.Node {
	if len(steps) <= 1 {
		return start
	}
	return steps[len(steps)-2].node
}

type move struct {
	to     *maze.Node
	edge   *maze.Edge
	wall   bool
	dx, dy int
}

func (a adder) pick(cur, dest, forbid *maze.Node, search, allowed map[*maze.Node]bool) (move, bool) {
	var toward, away []move
	consider := func(e *maze.Edge, wall bool) {
		to := e.Other(cur)
		if to == forbid || search[to] || (to.OnSolution && to != dest) || !allowed[to] {
			return
		}
		mv := move{to: to, edge: e, wall: wall, dx: to.X - cur.X, dy: to.Y - cur.Y}
		if inBox(to, cur, dest) {
			toward = append(toward, mv)
		} else {
			away = append(away, mv)
		}
	}
	for _, e := range cur.Walls {
		consider(e, true)
	}
	for _, e := range cur.Connections {
		consider(e, false)
	}
	if len(toward)+len(away) == 0 {
		return move{}, false
	}
	if a.r.IntN(4) != 0 {
		if len(toward) > 0 {
			return toward[a.r.IntN(len(toward))], true
		}
		return away[a.r.IntN(len(away))], true
	}
	if len(away) > 0 {
		return away[a.r.IntN(len(away))], true
	}
	return toward[a.r.IntN(len(toward))], true
}

func mark(n *maze.Node, id maze.PathID) {
	n.OnSolution = true
	n.Path = id
}

func unmark(n *maze.Node) {
	n.OnSolution = false
	n.Path = maze.PathID{}
	n.Index = 0
}

func inBox(n, a, b *maze.Node) bool {
	return between(n.X, a.X, b.X) && between(n.Y, a.Y, b.Y)
}

func between(v, a, b int) bool {
	if a > b {
		a, b = b, a
	}
	return v >= a && v <= b
}
