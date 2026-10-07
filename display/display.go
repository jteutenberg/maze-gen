// Package display renders a maze as one character per tile.
package display

import (
	"strings"

	"github.com/jteutenberg/maze-gen/maze"
)

const (
	wallN = 1 << iota
	wallE
	wallS
	wallW
)

type renderer struct{}

// New returns a Renderer. y increases upward, so the first row printed is
// the top of the maze. Exits are bright yellow. A bridge overwrites the
// tile it shares a position with.
func New() maze.Renderer {
	return renderer{}
}

func (renderer) Render(m *maze.Maze) string {
	exits := map[*maze.Node]bool{}
	for _, p := range m.ExitPairs() {
		exits[p.A] = true
		exits[p.B] = true
	}
	var b strings.Builder
	for y := m.Size - 1; y >= 0; y-- {
		for x := 0; x < m.Size; x++ {
			b.WriteString(cell(m, x, y, exits))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func cell(m *maze.Maze, x, y int, exits map[*maze.Node]bool) string {
	at := m.NodesAt(x, y)
	if len(at) == 0 {
		return " "
	}
	base := at[0]
	g, inverse := glyph(walls(m, base))
	if bridge := bridgeAt(at); bridge != nil {
		g = bridgeGlyph(bridge)
		inverse = false
	}
	switch {
	case exits[base]:
		return "\033[1;33m" + g + "\033[0m"
	case inverse:
		return "\033[40;37m" + g + "\033[0m"
	default:
		return g
	}
}

func walls(m *maze.Maze, n *maze.Node) int {
	mask := 0
	if blocked(m, n, 0, 1) {
		mask |= wallN
	}
	if blocked(m, n, 1, 0) {
		mask |= wallE
	}
	if blocked(m, n, 0, -1) {
		mask |= wallS
	}
	if blocked(m, n, -1, 0) {
		mask |= wallW
	}
	return mask
}

func blocked(m *maze.Maze, n *maze.Node, dx, dy int) bool {
	at := m.NodesAt(n.X+dx, n.Y+dy)
	if len(at) == 0 {
		return true
	}
	for _, other := range at {
		if n.ConnTo(other) != nil {
			return false
		}
	}
	return true
}

func bridgeAt(at []*maze.Node) *maze.Node {
	if len(at) < 2 {
		return nil
	}
	return at[len(at)-1]
}

func bridgeGlyph(n *maze.Node) string {
	if len(n.Connections) == 0 {
		return "╫"
	}
	other := n.Connections[0].Other(n)
	if other.X == n.X {
		return "╫"
	}
	return "╪"
}

func glyph(mask int) (string, bool) {
	switch mask {
	case 0:
		return " ", false
	case wallN:
		return "╦", false
	case wallS:
		return "╩", false
	case wallW:
		return "╠", false
	case wallE:
		return "╣", false
	case wallW | wallS:
		return "╚", false
	case wallE | wallS:
		return "╝", false
	case wallW | wallN:
		return "╔", false
	case wallE | wallN:
		return "╗", false
	case wallN | wallS:
		return "═", true
	case wallE | wallW:
		return "║", true
	case wallN | wallE | wallS | wallW:
		return "·", false
	case wallE | wallS | wallW:
		return "╨", false
	case wallN | wallS | wallW:
		return "╞", false
	case wallN | wallE | wallS:
		return "╡", false
	case wallN | wallE | wallW:
		return "╥", false

	default:
		return "█", false
	}
}
func glyph2(mask int) (string, bool) {
	switch mask {
	case 0:
		return " ", false
	case wallN:
		return "-", false
	case wallS:
		return "_", false
	case wallW:
		return "▏", false
	case wallE:
		return "▕", false
	case wallW | wallS:
		return "L", false
	case wallE | wallS:
		return ")", false
	case wallW | wallN:
		return "F", false
	case wallE | wallN:
		return "7", false
	case wallN | wallS:
		return "=", true
	case wallE | wallW:
		return "\"", true
	case wallN | wallE | wallS | wallW:
		return "·", false
	case wallE | wallS | wallW:
		return "U", false
	case wallN | wallS | wallW:
		return "C", false
	case wallN | wallE | wallS:
		return "3", false
	case wallN | wallE | wallW:
		return "N", false
	default:
		return "█", false
	}
}
