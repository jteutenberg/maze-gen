package maze

// Filler is step 6. It opens the remaining regions onto a solution
// and carves every node in them. Passages opened here stay open.
// Fill asks nothing.
type Filler interface {
	Fill(m *Maze) error
}
