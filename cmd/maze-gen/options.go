package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jteutenberg/maze-gen/maze"
)

const usage = `Usage: maze-gen [--size n] [--exits x,y x,y] [--exits-internal x,y x,y] [--rooms 0|1] [--dfs]

  --size n
        Side length from 5 to 30. If omitted, you are asked.
  --exits x,y x,y
        Two border exits. If omitted together with --exits-internal, you are
        asked and the positions are chosen at random.
  --exits-internal x,y x,y
        Two interior exits. If only one exit flag is set, the other kind of
        exit is left out. If both are set, the pairs are the first border exit
        with the first interior exit, and the second interior exit with the
        second border exit.
  --rooms 0|1
        Rooms to open. Defaults to 0.
  --dfs
        Open passages by a depth-first search from the first exit, instead of
        rooms, a solution, loops, and the fill. Rejected when both --exits
        and --exits-internal are set.
`

// errHelp asks main to print usage and exit successfully.
var errHelp = errors.New("help")

type coord struct {
	x, y int
}

// options is the part of the run the flags take over from stdin.
// A nil size means ask. Nil exit slices mean that kind of exit is off,
// and both nil means ask. Rooms are 0 unless --rooms says otherwise.
// dfs replaces rooms, the solution, loops, and the fill.
type options struct {
	size     *int
	border   []coord
	interior []coord
	rooms    int
	dfs      bool
}

func parseArgs(args []string) (options, error) {
	var opt options
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		name := args[i]
		if name == "--help" || name == "-h" {
			return options{}, errHelp
		}
		if !strings.HasPrefix(name, "--") {
			return options{}, fmt.Errorf("unexpected argument %q", name)
		}
		if seen[name] {
			return options{}, fmt.Errorf("%s given more than once", name)
		}
		seen[name] = true
		var err error
		switch name {
		case "--size":
			var n int
			n, i, err = takeInt(args, i, name)
			if err != nil {
				return options{}, err
			}
			if n < 5 || n > 30 {
				return options{}, fmt.Errorf("size %d is outside 5..30", n)
			}
			opt.size = &n
		case "--rooms":
			opt.rooms, i, err = takeInt(args, i, name)
			if err != nil {
				return options{}, err
			}
			if opt.rooms != 0 && opt.rooms != 1 {
				return options{}, fmt.Errorf("rooms %d is not 0 or 1", opt.rooms)
			}
		case "--exits":
			opt.border, i, err = takeCoords(args, i, name)
			if err != nil {
				return options{}, err
			}
		case "--exits-internal":
			opt.interior, i, err = takeCoords(args, i, name)
			if err != nil {
				return options{}, err
			}
		case "--dfs":
			opt.dfs = true
		default:
			return options{}, fmt.Errorf("unknown argument %q", name)
		}
	}
	if opt.dfs && opt.border != nil && opt.interior != nil {
		return options{}, fmt.Errorf("--dfs cannot be used with both --exits and --exits-internal")
	}
	if opt.dfs && opt.rooms != 0 {
		return options{}, fmt.Errorf("--dfs does not add rooms")
	}
	return opt, nil
}

func (opt options) sizeOrPrompt(in io.Reader) (int, error) {
	if opt.size != nil {
		return *opt.size, nil
	}
	fmt.Print("Maze size (5-30): ")
	var size int
	if _, err := fmt.Fscan(in, &size); err != nil {
		return 0, fmt.Errorf("read size: %w", err)
	}
	return size, nil
}

func (opt options) exitsChosen() bool {
	return opt.border != nil || opt.interior != nil
}

func (opt options) positions() (border, interior []maze.Pos) {
	return toPos(opt.border), toPos(opt.interior)
}

func toPos(cs []coord) []maze.Pos {
	if cs == nil {
		return nil
	}
	out := make([]maze.Pos, len(cs))
	for i, c := range cs {
		out[i] = maze.Pos{X: c.x, Y: c.y}
	}
	return out
}

func takeInt(args []string, i int, name string) (int, int, error) {
	if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
		return 0, i, fmt.Errorf("%s needs a value", name)
	}
	n, err := strconv.Atoi(args[i+1])
	if err != nil {
		return 0, i, fmt.Errorf("%s: %q is not an integer", name, args[i+1])
	}
	return n, i + 1, nil
}

func takeCoords(args []string, i int, name string) ([]coord, int, error) {
	if i+2 >= len(args) {
		return nil, i, fmt.Errorf("%s needs two coordinate pairs, as in %s 0,0 4,9", name, name)
	}
	cs := make([]coord, 2)
	for k := range cs {
		c, err := parseCoord(args[i+1+k])
		if err != nil {
			return nil, i, fmt.Errorf("%s: %w", name, err)
		}
		cs[k] = c
	}
	return cs, i + 2, nil
}

func parseCoord(s string) (coord, error) {
	body := strings.TrimSpace(s)
	body = strings.TrimPrefix(body, "(")
	body = strings.TrimSuffix(body, ")")
	xStr, yStr, ok := strings.Cut(body, ",")
	if !ok || strings.Contains(yStr, ",") {
		return coord{}, fmt.Errorf("%q is not a coordinate pair x,y", s)
	}
	x, errX := strconv.Atoi(strings.TrimSpace(xStr))
	y, errY := strconv.Atoi(strings.TrimSpace(yStr))
	if errX != nil || errY != nil {
		return coord{}, fmt.Errorf("%q is not a coordinate pair x,y", s)
	}
	return coord{x: x, y: y}, nil
}
