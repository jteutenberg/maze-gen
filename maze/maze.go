// Package maze holds the maze graph and the step interfaces that build it.
package maze

import "fmt"

// Pos is a position on the plane. It does not identify a node: a bridge
// shares its position with the tile it crosses.
type Pos struct {
	X, Y int
}

// Node is a tile. Identity is the pointer, not the position.
type Node struct {
	X, Y        int
	Walls       []*Edge
	Connections []*Edge
	Border      bool

	// OnSolution is set while a node is on a committed solution, including
	// the path currently being walked. Index counts from 0 at one exit.
	OnSolution bool
	Path       PathID
	Index      int
}

// Edge is an undirected link between two nodes. The same pointer sits in
// both nodes' wall lists, or in both connection lists once the wall is opened.
type Edge struct {
	A, B *Node
}

// Maze is the graph under construction.
type Maze struct {
	Size      int
	Nodes     []*Node
	at        map[Pos][]*Node
	exitPairs []ExitPair
	solutions []Solution
}

// NewMaze returns an empty maze that will hold a square of the given size.
func NewMaze(size int) *Maze {
	return &Maze{Size: size, at: make(map[Pos][]*Node)}
}

// AddNode appends a tile at (x, y). border marks a perimeter tile of the
// initial square. Later nodes at the same position, such as bridges, are
// stored after the original tile.
func (m *Maze) AddNode(x, y int, border bool) *Node {
	n := &Node{X: x, Y: y, Border: border}
	m.Nodes = append(m.Nodes, n)
	p := Pos{X: x, Y: y}
	m.at[p] = append(m.at[p], n)
	return n
}

// NodesAt returns every node at (x, y), original tile first.
func (m *Maze) NodesAt(x, y int) []*Node {
	return m.at[Pos{X: x, Y: y}]
}

// BorderNodes returns the perimeter tiles of the initial square, in the
// order they were added.
func (m *Maze) BorderNodes() []*Node {
	var out []*Node
	for _, n := range m.Nodes {
		if n.Border {
			out = append(out, n)
		}
	}
	return out
}

// AddWall links two orthogonally adjacent nodes with a closed edge.
func (m *Maze) AddWall(a, b *Node) (*Edge, error) {
	if !orthogonal(a, b) {
		return nil, fmt.Errorf("wall from (%d,%d) to (%d,%d) is not orthogonal", a.X, a.Y, b.X, b.Y)
	}
	e := &Edge{A: a, B: b}
	a.Walls = append(a.Walls, e)
	b.Walls = append(b.Walls, e)
	return e, nil
}

// Open turns a wall into a connection. The same edge pointer moves from
// both nodes' wall lists to their connection lists.
func (m *Maze) Open(e *Edge) error {
	if e == nil || e.A == nil || e.B == nil {
		return fmt.Errorf("nil edge")
	}
	if e.A.WallTo(e.B) != e || e.B.WallTo(e.A) != e {
		return fmt.Errorf("edge between (%d,%d) and (%d,%d) is not a wall", e.A.X, e.A.Y, e.B.X, e.B.Y)
	}
	removeEdge(&e.A.Walls, e)
	removeEdge(&e.B.Walls, e)
	e.A.Connections = append(e.A.Connections, e)
	e.B.Connections = append(e.B.Connections, e)
	return nil
}

// Close turns a connection back into a wall.
func (m *Maze) Close(e *Edge) error {
	if e == nil || e.A == nil || e.B == nil {
		return fmt.Errorf("nil edge")
	}
	if e.A.ConnTo(e.B) != e || e.B.ConnTo(e.A) != e {
		return fmt.Errorf("edge between (%d,%d) and (%d,%d) is not a connection", e.A.X, e.A.Y, e.B.X, e.B.Y)
	}
	removeEdge(&e.A.Connections, e)
	removeEdge(&e.B.Connections, e)
	e.A.Walls = append(e.A.Walls, e)
	e.B.Walls = append(e.B.Walls, e)
	return nil
}

// RemoveWall deletes a wall. The two nodes are left unlinked.
func (m *Maze) RemoveWall(e *Edge) error {
	if e == nil || e.A == nil || e.B == nil {
		return fmt.Errorf("nil edge")
	}
	if e.A.WallTo(e.B) != e || e.B.WallTo(e.A) != e {
		return fmt.Errorf("edge between (%d,%d) and (%d,%d) is not a wall", e.A.X, e.A.Y, e.B.X, e.B.Y)
	}
	removeEdge(&e.A.Walls, e)
	removeEdge(&e.B.Walls, e)
	return nil
}

// AddConnection links two orthogonally adjacent nodes with an open edge.
func (m *Maze) AddConnection(a, b *Node) (*Edge, error) {
	if !orthogonal(a, b) {
		return nil, fmt.Errorf("connection from (%d,%d) to (%d,%d) is not orthogonal", a.X, a.Y, b.X, b.Y)
	}
	e := &Edge{A: a, B: b}
	a.Connections = append(a.Connections, e)
	b.Connections = append(b.Connections, e)
	return e, nil
}

// Disconnect deletes a connection. The two nodes are left unlinked.
func (m *Maze) Disconnect(e *Edge) error {
	if e == nil || e.A == nil || e.B == nil {
		return fmt.Errorf("nil edge")
	}
	if e.A.ConnTo(e.B) != e || e.B.ConnTo(e.A) != e {
		return fmt.Errorf("edge between (%d,%d) and (%d,%d) is not a connection", e.A.X, e.A.Y, e.B.X, e.B.Y)
	}
	removeEdge(&e.A.Connections, e)
	removeEdge(&e.B.Connections, e)
	return nil
}

// RemoveNode drops a node that has no remaining edges. Used when a bridge
// is uncarved. The original tile at that position is left in place.
func (m *Maze) RemoveNode(n *Node) error {
	if n == nil {
		return fmt.Errorf("nil node")
	}
	if len(n.Walls) != 0 || len(n.Connections) != 0 {
		return fmt.Errorf("node (%d,%d) still has edges", n.X, n.Y)
	}
	p := Pos{X: n.X, Y: n.Y}
	at := m.at[p]
	found := false
	for i, cand := range at {
		if cand == n {
			m.at[p] = append(at[:i], at[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("node (%d,%d) is not in the maze", n.X, n.Y)
	}
	if len(m.at[p]) == 0 {
		delete(m.at, p)
	}
	for i, cand := range m.Nodes {
		if cand == n {
			m.Nodes = append(m.Nodes[:i], m.Nodes[i+1:]...)
			break
		}
	}
	return nil
}

// Other returns the endpoint that is not n.
func (e *Edge) Other(n *Node) *Node {
	if e.A == n {
		return e.B
	}
	return e.A
}

// WallTo returns the closed edge from n to other, or nil.
func (n *Node) WallTo(other *Node) *Edge {
	return findEdge(n.Walls, n, other)
}

// ConnTo returns the open edge from n to other, or nil.
func (n *Node) ConnTo(other *Node) *Edge {
	return findEdge(n.Connections, n, other)
}

func removeEdge(list *[]*Edge, e *Edge) bool {
	edges := *list
	for i, cur := range edges {
		if cur == e {
			*list = append(edges[:i], edges[i+1:]...)
			return true
		}
	}
	return false
}

func findEdge(edges []*Edge, from, to *Node) *Edge {
	for _, e := range edges {
		if e.Other(from) == to {
			return e
		}
	}
	return nil
}

func orthogonal(a, b *Node) bool {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx+dy == 1
}

// Adjacent reports whether a and b are orthogonal neighbours.
// Diagonal nodes are not adjacent.
func Adjacent(a, b *Node) bool {
	if a == nil || b == nil {
		return false
	}
	return orthogonal(a, b)
}
