// Command maze-gen walks the maze-building steps in order.
package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/jteutenberg/maze-gen/display"
	"github.com/jteutenberg/maze-gen/exits"
	"github.com/jteutenberg/maze-gen/fill"
	"github.com/jteutenberg/maze-gen/loops"
	"github.com/jteutenberg/maze-gen/maze"
	"github.com/jteutenberg/maze-gen/rooms"
	"github.com/jteutenberg/maze-gen/solution"
	"github.com/jteutenberg/maze-gen/square"
)

func main() {
	m, err := run()
	if err != nil {
		if m != nil {
			fmt.Fprint(os.Stderr, display.New().Render(m))
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (*maze.Maze, error) {
	fmt.Print("Maze size (5-30): ")
	var size int
	if _, err := fmt.Scan(&size); err != nil {
		return nil, fmt.Errorf("read size: %w", err)
	}
	topology, err := square.New(size)
	if err != nil {
		return nil, err
	}
	m, err := topology.Build()
	if err != nil {
		return nil, err
	}
	walls := 0
	for _, n := range m.Nodes {
		walls += len(n.Walls)
	}
	fmt.Printf("Topology: %d×%d, %d nodes, %d walls, %d border nodes\n",
		m.Size, m.Size, len(m.Nodes), walls/2, len(m.BorderNodes()))

	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1))
	if err := chooseExits(m, r); err != nil {
		return m, err
	}
	if err := chooseRooms(m, r); err != nil {
		return m, err
	}
	if err := buildSolutions(m, r); err != nil {
		return m, err
	}
	if err := addLoops(m, r); err != nil {
		return m, err
	}
	if err := fill.New(r).Fill(m); err != nil {
		return m, err
	}
	fmt.Print(display.New().Render(m))
	return m, nil
}

func chooseExits(m *maze.Maze, r *rand.Rand) error {
	fmt.Println("Exit placement:")
	fmt.Println("  1  both on the border")
	fmt.Println("  2  both interior")
	fmt.Println("  3  two pairs (border-interior and interior-border)")
	fmt.Print("Choice: ")
	var choice int
	if _, err := fmt.Scan(&choice); err != nil {
		return fmt.Errorf("read exit choice: %w", err)
	}
	var assigner maze.ExitAssigner
	switch choice {
	case 1:
		assigner = exits.NewBorder(r)
	case 2:
		assigner = exits.NewInterior(r)
	case 3:
		assigner = exits.NewMixed(r)
	default:
		return fmt.Errorf("exit choice %d is not 1, 2, or 3", choice)
	}
	if err := assigner.Assign(m); err != nil {
		return err
	}
	for i, p := range m.ExitPairs() {
		fmt.Printf("Exit pair %d: (%d,%d)-(%d,%d)\n", i+1, p.A.X, p.A.Y, p.B.X, p.B.Y)
	}
	return nil
}

func chooseRooms(m *maze.Maze, r *rand.Rand) error {
	adder := rooms.New(r)
	max := adder.MaxRooms(m)
	fmt.Printf("Rooms (0-%d): ", max)
	var n int
	if _, err := fmt.Scan(&n); err != nil {
		return fmt.Errorf("read room count: %w", err)
	}
	if err := adder.AddRooms(m, n); err != nil {
		return err
	}
	fmt.Printf("Opened %d rooms\n", n)
	return nil
}

func buildSolutions(m *maze.Maze, r *rand.Rand) error {
	if err := solution.New(r).Build(m); err != nil {
		return err
	}
	for i, s := range m.Solutions() {
		bridges := 0
		for _, n := range s.Nodes {
			at := m.NodesAt(n.X, n.Y)
			if len(at) > 0 && at[0] != n {
				bridges++
			}
		}
		end := s.Nodes[len(s.Nodes)-1]
		fmt.Printf("Solution %d: %d nodes, %d bridges, (%d,%d)-(%d,%d)\n",
			i+1, len(s.Nodes), bridges, s.Nodes[0].X, s.Nodes[0].Y, end.X, end.Y)
	}
	return nil
}

func addLoops(m *maze.Maze, r *rand.Rand) error {
	before := len(m.Solutions())
	if err := loops.New(r).AddLoops(m); err != nil {
		return err
	}
	fmt.Printf("Loops: %d\n", len(m.Solutions())-before)
	return nil
}
