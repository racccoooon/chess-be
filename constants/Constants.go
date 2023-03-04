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

	NonSpecialMove = 0
	EnPassant      = 1
	Castling       = 2
	Promotion      = 3

	IsNotCheck  = 0
	IsCheck     = 1
	IsCheckmate = 2
	IsStalemate = 3
)

func StatusAsString(status int) string {
	switch status {
	case IsNotCheck:
		return "isNotCheck"
	case IsCheck:
		return "isCheck"
	case IsCheckmate:
		return "isCheckmate"
	case IsStalemate:
		return "isStalemate"
	}

	panic("invalid status")
}

func MoveKindAsString(kind int) string {
	switch kind {
	case NonSpecialMove:
		return "nonSpecialMove"
	case EnPassant:
		return "enPassant"
	case Castling:
		return "castling"
	case Promotion:
		return "promotion"
	}

	panic("invalid move kind")
}

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

func PromotionTypeFromString(promotionType string) int {
	switch promotionType {
	case "queen":
		return Queen
	case "rook":
		return Rook
	case "knight":
		return Knight
	case "bishop":
		return Bishop
	}

	panic("invalid promotion type")
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

func TypeFromString(t string) int {
	switch t {
	case "pawn":
		return Pawn
	case "rook":
		return Rook
	case "knight":
		return Knight
	case "bishop":
		return Bishop
	case "queen":
		return Queen
	case "king":
		return King
	}

	panic("invalid type")
}

func GetOppositeColor(color int) int {
	if color == White {
		return Black
	}

	return White
}
