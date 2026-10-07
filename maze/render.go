package maze

// Renderer draws a finished maze. One character per tile, with exits coloured.
type Renderer interface {
	Render(m *Maze) string
}
