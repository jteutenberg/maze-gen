package maze

// Topology is step 1. It lays down the initial square: one node per tile,
// walls between orthogonal neighbours, and no connections.
//
// The command-line app collects the size and constructs an implementation.
// Build itself takes no input, so another topology can replace this one
// without changing the call.
type Topology interface {
	Build() (*Maze, error)
}
