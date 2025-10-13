package parser

// Position represents a location in the YAML file (line and column numbers).
type Position struct {
	Line   int // Line number (1-based)
	Column int // Column number (1-based)
}

// PositionMap maps YAML paths to their positions in the file.
type PositionMap map[string]Position

// NewPosition creates a new position with the given line and column.
func NewPosition(line, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

// NewPositionMap creates a new empty position map.
func NewPositionMap() PositionMap {
	return make(PositionMap)
}

// Set adds a position for the given path.
func (pm PositionMap) Set(path string, pos Position) {
	pm[path] = pos
}

// Get retrieves the position for the given path.
func (pm PositionMap) Get(path string) (Position, bool) {
	pos, ok := pm[path]
	return pos, ok
}

// GetOrDefault retrieves the position for the given path, or returns a zero position if not found.
func (pm PositionMap) GetOrDefault(path string) Position {
	if pos, ok := pm[path]; ok {
		return pos
	}
	return Position{Line: 0, Column: 0}
}
