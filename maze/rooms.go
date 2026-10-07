package maze

// RoomAdder is step 3. MaxRooms tells the command-line app how many rooms
// it may ask for. AddRooms opens that many non-overlapping 2×2 blocks.
type RoomAdder interface {
	MaxRooms(m *Maze) int
	AddRooms(m *Maze, n int) error
}
