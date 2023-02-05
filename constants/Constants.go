package constants

const (
	White       = 0
	Black       = 1
	RandomColor = 2

	Pawn   = 0
	Rook   = 1
	Knight = 2
	Bishop = 3
	Queen  = 4
	King   = 5
)

func ColorFromString(color string) int {
	switch color {
	case "white":
		return White
	case "black":
		return Black
	case "randomColor":
		return RandomColor
	}

	panic("invalid color")
}

func ColorAsString(color int) string {
	switch color {
	case White:
		return "white"
	case Black:
		return "black"
	}

	panic("invalid color")
}

func TypeAsString(t int) string {
	switch t {
	case Pawn:
		return "pawn"
	case Rook:
		return "rook"
	case Knight:
		return "knight"
	case Bishop:
		return "bishop"
	case Queen:
		return "queen"
	case King:
		return "king"
	}

	panic("invalid type")
}
