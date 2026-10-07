package maze

// LoopAdder is step 5. It carves loops through regions that touch a
// solution in two places far enough apart. Regions that cannot take a
// loop are left incomplete for the fill step. AddLoops asks nothing.
type LoopAdder interface {
	AddLoops(m *Maze) error
}
