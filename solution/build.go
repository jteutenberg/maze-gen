// Package solution joins each exit pair with a path through two waypoints.
package solution

import (
	"fmt"
	"math/rand/v2"

	"github.com/jteutenberg/maze-gen/maze"
)

const (
	kindWalk = iota
	kindBridge
	maxBridges  = 2
	maxAttempts = 40
)

type builder struct {
	r *rand.Rand
}

// New returns a SolutionBuilder. Abandoned branches are uncarved, and each
// solution is steered through two waypoints that are also far from each other.
func New(r *rand.Rand) maze.SolutionBuilder {
	return builder{r: r}
}

func (b builder) Build(m *maze.Maze) error {
	pairs := m.ExitPairs()
	if len(pairs) == 0 {
		return fmt.Errorf("no exit pairs")
	}
	for _, pair := range pairs {
		if err := b.buildPair(m, pair); err != nil {
			return err
		}
	}
	return nil
}

func (b builder) buildPair(m *maze.Maze, pair maze.ExitPair) error {
	id := maze.PathIDFor(pair.A, pair.B)
	var last error
	// Two waypoints, then one, then a direct walk. Each failure drops a
	// waypoint and tries the next simpler route.
	for _, attempt := range []func() error{
		func() error { return b.withTwoWaypoints(m, pair, id) },
		func() error { return b.withOneWaypoint(m, pair, id) },
		func() error { return b.withDirect(m, pair, id) },
	} {
		if err := attempt(); err != nil {
			last = err
			continue
		}
		return nil
	}
	if last == nil {
		last = fmt.Errorf("could not join exits (%d,%d) and (%d,%d)", pair.A.X, pair.A.Y, pair.B.X, pair.B.Y)
	}
	return last
}

func (b builder) withTwoWaypoints(m *maze.Maze, pair maze.ExitPair, id maze.PathID) error {
	var last error
	for range maxAttempts {
		w1, w2, err := b.pickWaypoints(m, pair)
		if err != nil {
			return err
		}
		legs, err := b.walkPair(m, pair, w1, w2, id)
		if err != nil {
			last = err
			continue
		}
		return commit(m, id, joinLegs(legs), [2]*maze.Node{w1, w2})
	}
	if last == nil {
		last = fmt.Errorf("could not join exits (%d,%d) and (%d,%d)", pair.A.X, pair.A.Y, pair.B.X, pair.B.Y)
	}
	return last
}

func (b builder) withOneWaypoint(m *maze.Maze, pair maze.ExitPair, id maze.PathID) error {
	cands := b.waypointCandidates(m, pair)
	if len(cands) == 0 {
		return fmt.Errorf("no waypoint candidates")
	}
	b.r.Shuffle(len(cands), func(i, j int) {
		cands[i], cands[j] = cands[j], cands[i]
	})
	if len(cands) > maxAttempts {
		cands = cands[:maxAttempts]
	}
	var last error
	for _, w := range cands {
		legs, err := b.walkToWaypoint(m, pair, w, id)
		if err != nil {
			last = err
			continue
		}
		return commit(m, id, joinAtWaypoint(legs[0], legs[1]), [2]*maze.Node{w, nil})
	}
	return last
}

func (b builder) withDirect(m *maze.Maze, pair maze.ExitPair, id maze.PathID) error {
	bridges := 0
	lg, err := b.walk(m, pair.A, pair.B, id, &bridges, true, true)
	if err != nil {
		return err
	}
	return commit(m, id, lg.nodes(), [2]*maze.Node{})
}

func commit(m *maze.Maze, id maze.PathID, seq []*maze.Node, waypoints [2]*maze.Node) error {
	for i, n := range seq {
		n.OnSolution = true
		n.Path = id
		n.Index = i
	}
	if err := pathConnected(seq); err != nil {
		return err
	}
	m.AddSolution(maze.Solution{
		ID:        id,
		Nodes:     seq,
		Waypoints: waypoints,
	})
	return nil
}

// walkPair walks exit A to w1, exit B to w2, then w1 to w2.
// On failure every leg of this attempt is uncarved.
func (b builder) walkPair(m *maze.Maze, pair maze.ExitPair, w1, w2 *maze.Node, id maze.PathID) ([]leg, error) {
	bridges := 0
	// The walks that leave an exit bend within two steps when they can, so the
	// exit is not visible down a straight corridor. The waypoint join does not.
	targets := []struct {
		from, to *maze.Node
		hideExit bool
	}{
		{pair.A, w1, true},
		{pair.B, w2, true},
		{w1, w2, false},
	}
	legs := make([]leg, 0, 3)
	for _, t := range targets {
		lg, err := b.walk(m, t.from, t.to, id, &bridges, t.hideExit, false)
		if err != nil {
			for i := len(legs) - 1; i >= 0; i-- {
				legs[i].undo(m)
			}
			return nil, err
		}
		legs = append(legs, lg)
	}
	return legs, nil
}

// walkToWaypoint walks each exit to the same waypoint. Both walks leave
// their exit, so each end can bend. On failure the attempt is uncarved.
func (b builder) walkToWaypoint(m *maze.Maze, pair maze.ExitPair, w *maze.Node, id maze.PathID) ([2]leg, error) {
	bridges := 0
	var legs [2]leg
	first, err := b.walk(m, pair.A, w, id, &bridges, true, false)
	if err != nil {
		return legs, err
	}
	second, err := b.walk(m, pair.B, w, id, &bridges, true, false)
	if err != nil {
		first.undo(m)
		return legs, err
	}
	legs[0] = first
	legs[1] = second
	return legs, nil
}

func joinAtWaypoint(toW, fromOtherExit leg) []*maze.Node {
	seq := append([]*maze.Node{}, toW.nodes()...)
	back := fromOtherExit.nodes()
	for i := len(back) - 2; i >= 0; i-- {
		seq = append(seq, back[i])
	}
	return seq
}

func joinLegs(legs []leg) []*maze.Node {
	a := legs[0].nodes()
	b := legs[1].nodes()
	join := legs[2].nodes()
	seq := append([]*maze.Node{}, a...)
	if len(join) > 0 {
		seq = append(seq, join[1:]...)
	}
	for i := len(b) - 2; i >= 0; i-- {
		seq = append(seq, b[i])
	}
	return seq
}

func pathConnected(nodes []*maze.Node) error {
	if len(nodes) < 2 {
		return fmt.Errorf("solution has %d nodes", len(nodes))
	}
	for i := 0; i < len(nodes)-1; i++ {
		if nodes[i].ConnTo(nodes[i+1]) == nil {
			return fmt.Errorf("solution gap between (%d,%d) and (%d,%d)", nodes[i].X, nodes[i].Y, nodes[i+1].X, nodes[i+1].Y)
		}
	}
	return nil
}

func (b builder) waypointCandidates(m *maze.Maze, pair maze.ExitPair) []*maze.Node {
	minDist := m.Size / 2
	exits := map[*maze.Node]bool{}
	for _, p := range m.ExitPairs() {
		exits[p.A] = true
		exits[p.B] = true
	}
	var cands []*maze.Node
	for _, n := range m.Nodes {
		if exits[n] || n.OnSolution {
			continue
		}
		if manhattan(n, pair.A) < minDist || manhattan(n, pair.B) < minDist {
			continue
		}
		cands = append(cands, n)
	}
	return cands
}

func (b builder) pickWaypoints(m *maze.Maze, pair maze.ExitPair) (*maze.Node, *maze.Node, error) {
	minDist := m.Size / 2
	cands := b.waypointCandidates(m, pair)
	if len(cands) < 2 {
		return nil, nil, fmt.Errorf("only %d waypoint candidates", len(cands))
	}
	b.r.Shuffle(len(cands), func(i, j int) {
		cands[i], cands[j] = cands[j], cands[i]
	})
	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			if manhattan(cands[i], cands[j]) >= minDist {
				return cands[i], cands[j], nil
			}
		}
	}
	return nil, nil, fmt.Errorf("no waypoint pair is far enough apart")
}

type leg struct {
	start       *maze.Node
	markedStart bool
	steps       []step
}

func (l leg) nodes() []*maze.Node {
	out := []*maze.Node{l.start}
	for _, s := range l.steps {
		if s.kind == kindBridge {
			out = append(out, s.bridge)
		}
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
		undoStep(m, prev, l.steps[i])
	}
	if l.markedStart {
		unmark(l.start)
	}
}

type step struct {
	kind   int
	node   *maze.Node
	dx, dy int
	marked bool
	edge   *maze.Edge
	opened bool
	bridge *maze.Node
	linkA  *maze.Edge
	linkB  *maze.Edge
	mid    *maze.Node
}

func (b builder) walk(m *maze.Maze, start, dest *maze.Node, id maze.PathID, bridges *int, hideExit, bendAtDest bool) (leg, error) {
	lg := leg{start: start}
	if !start.OnSolution {
		mark(start, id)
		lg.markedStart = true
	}
	search := map[*maze.Node]bool{start: true}
	guard := 0
	limit := len(m.Nodes)*8 + 64
	for {
		guard++
		if guard > limit {
			lg.undo(m)
			return leg{}, fmt.Errorf("solution search exceeded %d steps", limit)
		}
		cur := start
		if len(lg.steps) > 0 {
			cur = lg.steps[len(lg.steps)-1].node
		}
		if cur == dest {
			return lg, nil
		}
		mv, ok := b.pick(m, cur, dest, lg.steps, search, *bridges, hideExit, bendAtDest)
		if !ok {
			if len(lg.steps) == 0 {
				lg.undo(m)
				return leg{}, fmt.Errorf("no path from (%d,%d) to (%d,%d)", start.X, start.Y, dest.X, dest.Y)
			}
			prev := start
			if len(lg.steps) > 1 {
				prev = lg.steps[len(lg.steps)-2].node
			}
			last := lg.steps[len(lg.steps)-1]
			if last.kind == kindBridge {
				*bridges--
			}
			undoStep(m, prev, last)
			lg.steps = lg.steps[:len(lg.steps)-1]
			continue
		}
		st, err := applyStep(m, cur, mv, id)
		if err != nil {
			lg.undo(m)
			return leg{}, err
		}
		if st.kind == kindBridge {
			*bridges++
			search[st.bridge] = true
		}
		search[st.node] = true
		lg.steps = append(lg.steps, st)
	}
}

type move struct {
	kind   int
	to     *maze.Node
	mid    *maze.Node
	edge   *maze.Edge
	opened bool
	dx, dy int
}

func (b builder) pick(m *maze.Maze, cur, dest *maze.Node, steps []step, search map[*maze.Node]bool, bridges int, hideExit, bendAtDest bool) (move, bool) {
	var bridgeOpts []move
	if bridges < maxBridges {
		bridgeOpts = bridgeMoves(m, cur, dest, search)
	}
	cands := stepMoves(m, cur, dest, search)
	// A bridge from the exit would place two collinear nodes immediately.
	// After one step, prefer any turn so the exit and its two nearest path
	// nodes are not in a straight line. Straight is used only when no turn exists.
	if hideExit && len(steps) == 0 && len(cands) > 0 {
		bridgeOpts = nil
	}
	if hideExit && len(steps) == 1 {
		bentBridges := notStraight(bridgeOpts, steps[0].dx, steps[0].dy)
		bentSteps := notStraight(cands, steps[0].dx, steps[0].dy)
		if len(bentBridges) > 0 || len(bentSteps) > 0 {
			bridgeOpts = bentBridges
			cands = bentSteps
		}
	}
	if bendAtDest && len(steps) >= 1 {
		dx, dy := steps[len(steps)-1].dx, steps[len(steps)-1].dy
		bridgeOpts, cands = deferStraightInto(bridgeOpts, cands, dest, dx, dy)
	}
	if len(bridgeOpts) > 0 {
		return bridgeOpts[b.r.IntN(len(bridgeOpts))], true
	}
	if len(cands) == 0 {
		return move{}, false
	}
	dx, dy, run := runLen(steps)
	if run >= m.Size/2 && (dx != 0 || dy != 0) {
		var turns []move
		lx, ly := -dy, dx
		rx, ry := dy, -dx
		for _, c := range cands {
			if (c.dx == lx && c.dy == ly) || (c.dx == rx && c.dy == ry) {
				turns = append(turns, c)
			}
		}
		if len(turns) > 0 {
			return turns[b.r.IntN(len(turns))], true
		}
	}
	var toward, away []move
	for _, c := range cands {
		if inBox(c.to, cur, dest) {
			toward = append(toward, c)
		} else {
			away = append(away, c)
		}
	}
	if b.r.IntN(4) != 0 {
		if len(toward) > 0 {
			return toward[b.r.IntN(len(toward))], true
		}
		return away[b.r.IntN(len(away))], true
	}
	if len(away) > 0 {
		return away[b.r.IntN(len(away))], true
	}
	return toward[b.r.IntN(len(toward))], true
}

func bridgeMoves(m *maze.Maze, cur, dest *maze.Node, search map[*maze.Node]bool) []move {
	var out []move
	for _, e := range cur.Walls {
		mid := e.Other(cur)
		if !mid.OnSolution || mid == dest || !inBox(mid, cur, dest) {
			continue
		}
		dx, dy := mid.X-cur.X, mid.Y-cur.Y
		at := m.NodesAt(mid.X+dx, mid.Y+dy)
		if len(at) == 0 {
			continue
		}
		far := at[0]
		if far == cur || far.OnSolution || search[far] || mid.WallTo(far) == nil {
			continue
		}
		out = append(out, move{kind: kindBridge, to: far, mid: mid, dx: dx, dy: dy})
	}
	return out
}

func stepMoves(m *maze.Maze, cur, dest *maze.Node, search map[*maze.Node]bool) []move {
	var out []move
	seen := map[*maze.Node]bool{}
	consider := func(e *maze.Edge, open bool) {
		to := e.Other(cur)
		if seen[to] || search[to] {
			return
		}
		if to.OnSolution && to != dest {
			return
		}
		seen[to] = true
		out = append(out, move{
			kind:   kindWalk,
			to:     to,
			edge:   e,
			opened: !open,
			dx:     to.X - cur.X,
			dy:     to.Y - cur.Y,
		})
	}
	for _, e := range cur.Walls {
		consider(e, false)
	}
	for _, e := range cur.Connections {
		consider(e, true)
	}
	return out
}

func applyStep(m *maze.Maze, cur *maze.Node, mv move, id maze.PathID) (step, error) {
	if mv.kind == kindBridge {
		return applyBridge(m, cur, mv.mid, mv.to, id, mv.dx, mv.dy)
	}
	st := step{kind: kindWalk, node: mv.to, dx: mv.dx, dy: mv.dy, edge: mv.edge}
	if mv.opened {
		if err := m.Open(mv.edge); err != nil {
			return step{}, err
		}
		st.opened = true
	}
	if !mv.to.OnSolution {
		mark(mv.to, id)
		st.marked = true
	}
	return st, nil
}

func applyBridge(m *maze.Maze, cur, mid, far *maze.Node, id maze.PathID, dx, dy int) (step, error) {
	wallA := cur.WallTo(mid)
	wallB := mid.WallTo(far)
	if wallA == nil || wallB == nil {
		return step{}, fmt.Errorf("bridge at (%d,%d) is missing a wall", mid.X, mid.Y)
	}
	if err := m.RemoveWall(wallA); err != nil {
		return step{}, err
	}
	if err := m.RemoveWall(wallB); err != nil {
		return step{}, err
	}
	bridge := m.AddNode(mid.X, mid.Y, false)
	linkA, err := m.AddConnection(cur, bridge)
	if err != nil {
		return step{}, err
	}
	linkB, err := m.AddConnection(bridge, far)
	if err != nil {
		return step{}, err
	}
	mark(bridge, id)
	mark(far, id)
	return step{
		kind:   kindBridge,
		node:   far,
		dx:     dx,
		dy:     dy,
		marked: true,
		bridge: bridge,
		linkA:  linkA,
		linkB:  linkB,
		mid:    mid,
	}, nil
}

func undoStep(m *maze.Maze, prev *maze.Node, st step) {
	if st.kind == kindBridge {
		_ = m.Disconnect(st.linkA)
		_ = m.Disconnect(st.linkB)
		unmark(st.bridge)
		_ = m.RemoveNode(st.bridge)
		_, _ = m.AddWall(prev, st.mid)
		_, _ = m.AddWall(st.mid, st.node)
		if st.marked {
			unmark(st.node)
		}
		return
	}
	if st.opened {
		_ = m.Close(st.edge)
	}
	if st.marked {
		unmark(st.node)
	}
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

// deferStraightInto drops a straight step onto dest while any other move
// remains, so the path can turn before it reaches that exit.
func deferStraightInto(bridges, cands []move, dest *maze.Node, dx, dy int) ([]move, []move) {
	straightIn := func(mv move) bool {
		return mv.to == dest && mv.dx == dx && mv.dy == dy
	}
	alts := 0
	for _, mv := range bridges {
		if !straightIn(mv) {
			alts++
		}
	}
	for _, mv := range cands {
		if !straightIn(mv) {
			alts++
		}
	}
	if alts == 0 {
		return bridges, cands
	}
	return filterMoves(bridges, straightIn), filterMoves(cands, straightIn)
}

func filterMoves(moves []move, drop func(move) bool) []move {
	var out []move
	for _, mv := range moves {
		if !drop(mv) {
			out = append(out, mv)
		}
	}
	return out
}

func notStraight(moves []move, dx, dy int) []move {
	var out []move
	for _, mv := range moves {
		if mv.dx != dx || mv.dy != dy {
			out = append(out, mv)
		}
	}
	return out
}

func runLen(steps []step) (dx, dy, n int) {
	if len(steps) == 0 {
		return 0, 0, 0
	}
	dx, dy = steps[len(steps)-1].dx, steps[len(steps)-1].dy
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].dx != dx || steps[i].dy != dy {
			break
		}
		n++
	}
	return dx, dy, n
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

func manhattan(a, b *maze.Node) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
