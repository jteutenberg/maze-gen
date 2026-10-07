package display

import (
	"regexp"
	"strings"
	"testing"

	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/square"
)

func TestUncarvedMazeIsClosed(t *testing.T) {
	m := mustSquare(t, 5)
	closed, _ := glyph(wallN | wallE | wallS | wallW)
	got := stripANSI(New().Render(m))
	want := strings.Repeat(strings.Repeat(closed, m.Size)+"\n", m.Size)
	if got != want {
		t.Fatalf("got\n%s", got)
	}
}

func TestOpenRowRendersACorridor(t *testing.T) {
	m := mustSquare(t, 5)
	for x := 0; x < m.Size-1; x++ {
		a := m.NodesAt(x, 0)[0]
		b := m.NodesAt(x+1, 0)[0]
		if err := m.Open(a.WallTo(b)); err != nil {
			t.Fatal(err)
		}
	}
	raw := New().Render(m)
	lines := strings.Split(strings.TrimSuffix(stripANSI(raw), "\n"), "\n")
	var want strings.Builder
	for x := 0; x < m.Size; x++ {
		g, inverse := glyph(walls(m, m.NodesAt(x, 0)[0]))
		want.WriteString(g)
		if inverse && !strings.Contains(raw, "\033[40;37m"+g+"\033[0m") {
			t.Fatalf("corridor glyph %q should use a black background", g)
		}
	}
	if got := lines[len(lines)-1]; got != want.String() {
		t.Fatalf("bottom row %q, want %q", got, want.String())
	}
}

func TestExitIsColoured(t *testing.T) {
	m := mustSquare(t, 5)
	a := m.NodesAt(0, 0)[0]
	b := m.NodesAt(4, 4)[0]
	if err := m.SetExitPairs([]maze.ExitPair{{A: a, B: b}}); err != nil {
		t.Fatal(err)
	}
	raw := New().Render(m)
	for _, exit := range []*maze.Node{a, b} {
		g, _ := glyph(walls(m, exit))
		if !strings.Contains(raw, "\033[1;33m"+g+"\033[0m") {
			t.Fatalf("exit (%d,%d) glyph %q is not yellow", exit.X, exit.Y, g)
		}
	}
}

func TestBridgeOverwritesItsCell(t *testing.T) {
	m := mustSquare(t, 5)
	left := m.NodesAt(1, 2)[0]
	mid := m.NodesAt(2, 2)[0]
	right := m.NodesAt(3, 2)[0]
	_ = m.RemoveWall(left.WallTo(mid))
	_ = m.RemoveWall(mid.WallTo(right))
	bridge := m.AddNode(2, 2, false)
	if _, err := m.AddConnection(left, bridge); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AddConnection(bridge, right); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(stripANSI(New().Render(m)), "\n"), "\n")
	// y=2 is the third row from the top in a 5-high maze.
	row := []rune(lines[m.Size-1-2])
	got := string(row[2])
	if got != bridgeGlyph(bridge) {
		t.Fatalf("bridge cell %q, want %q", got, bridgeGlyph(bridge))
	}
	under, _ := glyph(walls(m, mid))
	if got == under {
		t.Fatal("bridge glyph matched the tile underneath")
	}
}

func stripANSI(s string) string {
	return regexp.MustCompile(`\033\[[0-9;]*m`).ReplaceAllString(s, "")
}

func mustSquare(t *testing.T, size int) *maze.Maze {
	t.Helper()
	top, err := square.New(size)
	if err != nil {
		t.Fatal(err)
	}
	m, err := top.Build()
	if err != nil {
		t.Fatal(err)
	}
	return m
}
