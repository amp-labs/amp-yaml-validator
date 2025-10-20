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
// If the exact path is not found, it walks up the path hierarchy to find the nearest parent position.
// For example, if "$.integrations[0].read.objects[0].schedule" is not found, it tries:
// - "$.integrations[0].read.objects[0]"
// - "$.integrations[0].read.objects"
// - "$.integrations[0].read"
// - etc.
func (pm PositionMap) GetOrDefault(path string) Position {
	if pos, ok := pm[path]; ok {
		return pos
	}

	// Walk up the path hierarchy to find nearest parent position
	currentPath := path

	for {
		// Find the last segment (after last dot or bracket)
		lastDot := -1
		lastBracket := -1

		for i := len(currentPath) - 1; i >= 0; i-- {
			if currentPath[i] == '.' && lastDot == -1 {
				lastDot = i
			}

			if currentPath[i] == '[' && lastBracket == -1 {
				lastBracket = i
			}

			if lastDot != -1 && lastBracket != -1 {
				break
			}
		}

		// Determine where to cut the path
		var cutPos int

		if lastDot == -1 && lastBracket == -1 {
			// No more parents to try
			break
		} else if lastDot > lastBracket {
			// Last segment is after a dot (e.g., ".schedule")
			cutPos = lastDot
		} else {
			// Last segment is an array index (e.g., "[0]")
			cutPos = lastBracket
		}

		// Try parent path
		currentPath = currentPath[:cutPos]
		if pos, ok := pm[currentPath]; ok {
			return pos
		}
	}

	return Position{Line: 0, Column: 0}
}
