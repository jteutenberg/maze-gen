package main

import (
	"bytes"
	"errors"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jteutenberg/maze-gen/maze"
)

func TestParseArgs(t *testing.T) {
	opt, err := parseArgs([]string{
		"--rooms", "1",
		"--size", "12",
		"--exits", "(0,0)", "11,4",
		"--exits-internal", "2,2", "8, 9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opt.size == nil || *opt.size != 12 || opt.rooms != 1 {
		t.Fatalf("size %v rooms %d", opt.size, opt.rooms)
	}
	if len(opt.border) != 2 || opt.border[0] != (coord{0, 0}) || opt.border[1] != (coord{11, 4}) {
		t.Fatalf("border = %v", opt.border)
	}
	if len(opt.interior) != 2 || opt.interior[0] != (coord{2, 2}) || opt.interior[1] != (coord{8, 9}) {
		t.Fatalf("interior = %v", opt.interior)
	}
}

func TestParseArgsDefaults(t *testing.T) {
	opt, err := parseArgs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if opt.size != nil || opt.border != nil || opt.interior != nil || opt.rooms != 0 || opt.dfs {
		t.Fatalf("defaults = %+v", opt)
	}
	if opt.exitsChosen() {
		t.Fatal("exits chosen with no flags")
	}
}

func TestParseArgsErrors(t *testing.T) {
	cases := [][]string{
		{"--rooms", "2"},
		{"--rooms", "-1"},
		{"--rooms"},
		{"--size"},
		{"--size", "4"},
		{"--size", "31"},
		{"--size", "x"},
		{"--exits", "0,0"},
		{"--exits", "0,0", "nope"},
		{"--exits", "0,0", "1,1", "--exits", "2,2", "3,3"},
		{"--unknown"},
		{"5"},
		{"--exits-internal", "1,1", "2,2", "--rooms", "1", "--rooms", "0"},
		{"--dfs", "--exits", "0,0", "4,0", "--exits-internal", "1,1", "2,2"},
		{"--dfs", "--rooms", "1"},
		{"--dfs", "--dfs"},
	}
	for _, c := range cases {
		if _, err := parseArgs(c); err == nil {
			t.Errorf("parseArgs(%q) succeeded", c)
		}
	}
}

func TestHelp(t *testing.T) {
	_, err := parseArgs([]string{"--size", "8", "--help"})
	if !errors.Is(err, errHelp) {
		t.Fatalf("err = %v", err)
	}
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) {
	return 0, errors.New("unexpected read from stdin")
}

func TestPrepareUsesFlagsAndSkipsStdin(t *testing.T) {
	restore := captureStdout(t)
	m, _, err := prepare(failReader{}, []string{
		"--size", "8",
		"--exits", "0,0", "7,7",
		"--exits-internal", "2,2", "5,5",
		"--rooms", "1",
	}, rand.New(rand.NewPCG(1, 1)))
	out := restore()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Maze size") || strings.Contains(out, "Exit placement") || strings.Contains(out, "Rooms (") {
		t.Fatalf("prompted despite flags:\n%s", out)
	}
	pairs := m.ExitPairs()
	if len(pairs) != 2 {
		t.Fatalf("pairs = %d", len(pairs))
	}
	if pairs[0].A != m.NodesAt(0, 0)[0] || pairs[0].B != m.NodesAt(2, 2)[0] {
		t.Fatalf("first pair = (%d,%d)-(%d,%d)", pairs[0].A.X, pairs[0].A.Y, pairs[0].B.X, pairs[0].B.Y)
	}
	if pairs[1].A != m.NodesAt(5, 5)[0] || pairs[1].B != m.NodesAt(7, 7)[0] {
		t.Fatalf("second pair = (%d,%d)-(%d,%d)", pairs[1].A.X, pairs[1].A.Y, pairs[1].B.X, pairs[1].B.Y)
	}
	if got := countRooms(m); got != 1 {
		t.Fatalf("rooms = %d", got)
	}
}

func TestPrepareOneExitFlagDisablesTheOther(t *testing.T) {
	restore := captureStdout(t)
	defer restore()
	m, _, err := prepare(failReader{}, []string{
		"--size", "8",
		"--exits-internal", "3,1", "6,4",
	}, rand.New(rand.NewPCG(2, 2)))
	if err != nil {
		t.Fatal(err)
	}
	pairs := m.ExitPairs()
	if len(pairs) != 1 || pairs[0].A.Border || pairs[0].B.Border {
		t.Fatalf("got %d pairs", len(pairs))
	}
	if pairs[0].A != m.NodesAt(3, 1)[0] || pairs[0].B != m.NodesAt(6, 4)[0] {
		t.Fatalf("pair = (%d,%d)-(%d,%d)", pairs[0].A.X, pairs[0].A.Y, pairs[0].B.X, pairs[0].B.Y)
	}
	if got := countRooms(m); got != 0 {
		t.Fatalf("rooms = %d, want the default 0", got)
	}
}

func TestPreparePromptsForSizeAndExitsButNotRooms(t *testing.T) {
	restore := captureStdout(t)
	m, _, err := prepare(strings.NewReader("6\n1\n"), nil, rand.New(rand.NewPCG(3, 4)))
	out := restore()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Maze size") || !strings.Contains(out, "Exit placement") {
		t.Fatalf("missing prompts:\n%s", out)
	}
	if strings.Contains(out, "Rooms (") {
		t.Fatalf("still asked for rooms:\n%s", out)
	}
	if m.Size != 6 || len(m.ExitPairs()) != 1 {
		t.Fatalf("size %d pairs %d", m.Size, len(m.ExitPairs()))
	}
	for _, p := range m.ExitPairs() {
		if !p.A.Border || !p.B.Border {
			t.Fatal("choice 1 did not place border exits")
		}
	}
	if got := countRooms(m); got != 0 {
		t.Fatalf("rooms = %d", got)
	}
}

func TestPreparePromptsOnlyForTheMissingPiece(t *testing.T) {
	restore := captureStdout(t)
	defer restore()
	r := rand.New(rand.NewPCG(5, 6))
	withSize, _, err := prepare(strings.NewReader("2\n"), []string{"--size", "7"}, r)
	if err != nil {
		t.Fatal(err)
	}
	if withSize.Size != 7 || len(withSize.ExitPairs()) != 1 {
		t.Fatalf("size %d pairs %d", withSize.Size, len(withSize.ExitPairs()))
	}
	if withSize.ExitPairs()[0].A.Border || withSize.ExitPairs()[0].B.Border {
		t.Fatal("choice 2 did not place interior exits")
	}

	withExits, _, err := prepare(strings.NewReader("8\n"), []string{"--exits", "0,0", "0,7"}, rand.New(rand.NewPCG(7, 8)))
	if err != nil {
		t.Fatal(err)
	}
	p := withExits.ExitPairs()
	if withExits.Size != 8 || len(p) != 1 || p[0].A != withExits.NodesAt(0, 0)[0] || p[0].B != withExits.NodesAt(0, 7)[0] {
		t.Fatalf("size %d pairs %d", withExits.Size, len(p))
	}
}

func TestPrepareRejectsBadExplicitExits(t *testing.T) {
	restore := captureStdout(t)
	defer restore()
	_, _, err := prepare(failReader{}, []string{"--size", "8", "--exits", "1,1", "0,7"}, rand.New(rand.NewPCG(1, 1)))
	if err == nil {
		t.Fatal("expected an error")
	}
	_, _, err = prepare(failReader{}, []string{"--size", "5", "--exits", "0,0", "4,4", "--rooms", "1"}, rand.New(rand.NewPCG(1, 1)))
	if err == nil {
		t.Fatal("expected rooms to be rejected when the maze cannot hold one")
	}
}

func TestGenerateDFSSkipsLaterSteps(t *testing.T) {
	restore := captureStdout(t)
	m, err := generate(failReader{}, []string{
		"--size", "8",
		"--exits", "0,0", "0,7",
		"--dfs",
	}, rand.New(rand.NewPCG(4, 5)))
	out := restore()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "DFS from (0,0)") {
		t.Fatalf("output missing the dfs start:\n%s", out)
	}
	if strings.Contains(out, "Opened") || strings.Contains(out, "Solution") || strings.Contains(out, "Loops:") {
		t.Fatalf("dfs ran a later step:\n%s", out)
	}
	if len(m.Solutions()) != 0 {
		t.Fatalf("solutions = %d", len(m.Solutions()))
	}
	if len(m.Nodes) != 64 {
		t.Fatalf("nodes = %d", len(m.Nodes))
	}
	other := m.NodesAt(0, 7)[0]
	if len(other.Connections) != 1 {
		t.Fatalf("other exit has %d connections", len(other.Connections))
	}
}

func TestGenerateWithFlags(t *testing.T) {
	restore := captureStdout(t)
	m, err := generate(failReader{}, []string{
		"--size", "8",
		"--exits", "0,0", "0,7",
		"--rooms", "0",
	}, rand.New(rand.NewPCG(9, 10)))
	out := restore()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Exit pair 1: (0,0)-(0,7)") {
		t.Fatalf("output missing the fixed exits:\n%s", out)
	}
	if len(m.Solutions()) == 0 {
		t.Fatal("no solution")
	}
}

func countRooms(m *maze.Maze) int {
	n := 0
	for _, node := range m.Nodes {
		if len(node.Connections) == 2 {
			n++
		}
	}
	return n / 4
}

func captureStdout(t *testing.T) func() string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	var once sync.Once
	stop := func() string {
		once.Do(func() {
			w.Close()
			<-done
			os.Stdout = old
		})
		return buf.String()
	}
	t.Cleanup(func() { stop() })
	return stop
}
