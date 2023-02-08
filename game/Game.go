package game

import (
	"chess-be/constants"
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type Manager struct {
	games map[Id]*Game
}

func NewGameManager() *Manager {
	return &Manager{
		games: make(map[Id]*Game),
	}
}

func (g *Manager) NewGame(firstPlayerColor int) *Game {
	idValue, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	game := &Game{
		id:               Id(idValue.String()),
		firstPlayerColor: firstPlayerColor,

		turn: 0,

		players: make([]*Player, 0),
		pieces:  make([]Piece, 0),
		moves:   make([]Move, 0),
	}

	game.initializeBoard()

	g.games[game.id] = game

	return game
}

func (g *Manager) GetGame(id Id) *Game {
	if game, ok := g.games[id]; ok {
		return game
	}

	return nil
}

type Id string

type Game struct {
	id               Id
	firstPlayerColor int

	turn int

	players []*Player
	pieces  []Piece
	moves   []Move
}

type Move struct {
	color int
	t     int
	fromX int
	fromY int
	toX   int
	toY   int
	kind  int
}

func (g *Game) Pieces() []Piece {
	return g.pieces
}

func (Move *Move) Color() int {
	return Move.color
}

func (Move *Move) Type() int {
	return Move.t
}

func (Move *Move) Kind() int {
	return Move.kind
}

func (Move *Move) FromX() int {
	return Move.fromX
}

func (Move *Move) FromY() int {
	return Move.fromY
}

func (Move *Move) ToX() int {
	return Move.toX
}

func (Move *Move) ToY() int {
	return Move.toY
}

type Piece struct {
	color    int
	type_    int
	x        int
	y        int
	hasMoved bool
}

func (Piece *Piece) Color() int {
	return Piece.color
}

func (Piece *Piece) Type() int {
	return Piece.type_
}

func (Piece *Piece) X() int {
	return Piece.x
}

func (Piece *Piece) Y() int {
	return Piece.y
}

func (g *Game) Id() Id {
	return g.id
}

func (g *Game) ActiveColor() int {
	return g.turn % 2
}

func (g *Game) initializeBoard() {
	// white
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Rook, x: 0, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Knight, x: 1, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Bishop, x: 2, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Queen, x: 3, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.King, x: 4, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Bishop, x: 5, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Knight, x: 6, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Rook, x: 7, y: 0})

	for i := 0; i < 8; i++ {
		g.pieces = append(g.pieces, Piece{color: constants.White, type_: constants.Pawn, x: i, y: 1})
	}

	// black
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Rook, x: 0, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Knight, x: 1, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Bishop, x: 2, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Queen, x: 3, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.King, x: 4, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Bishop, x: 5, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Knight, x: 6, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Rook, x: 7, y: 7})

	for i := 0; i < 8; i++ {
		g.pieces = append(g.pieces, Piece{color: constants.Black, type_: constants.Pawn, x: i, y: 6})
	}
}

func (g *Game) LastMove() *Move {
	if len(g.moves) == 0 {
		return nil
	}

	return &g.moves[len(g.moves)-1]
}

func (g *Game) AddPlayer(name string, token string, connectionId string) *Player {
	color := g.firstPlayerColor
	if g.firstPlayerColor == constants.RandomColor {
		rand.Seed(time.Now().UnixNano())
		color = rand.Intn(2)
	}

	if len(g.players) >= 1 {
		color = g.players[0].color ^ 1
	}

	player := &Player{
		name:         name,
		token:        token,
		connectionId: connectionId,
		color:        color,
	}

	g.players = append(g.players, player)

	return player
}

func (g *Game) GetPlayerByToken(token string) *Player {
	for _, player := range g.players {
		if player.token == token {
			return player
		}
	}

	return nil
}

func (g *Game) GetPlayerByConnectionId(connectionId string) *Player {
	for _, player := range g.players {
		if player.connectionId == connectionId {
			return player
		}
	}

	return nil
}

func (g *Game) RejoinPlayer(token string, connectionId string) *Player {
	for _, player := range g.players {
		if player.token == token {
			player.connectionId = connectionId
		}
	}

	return nil
}

func (g *Game) OpponentName(color int) string {
	for _, player := range g.players {
		if player.color != color {
			return player.name
		}
	}

	return ""
}

func (g *Game) PlayerCount() int {
	return len(g.players)
}

func (g *Game) History() []Move {
	return g.moves
}

func (g *Game) Move(fromX int, fromY int, toX int, toY int) bool {
	piece := g.GetPieceAt(fromX, fromY)
	if piece == nil {
		return false
	}

	if piece.color != g.ActiveColor() {
		return false
	}

	isValidMove, moveType := g.IsMoveValid(*piece, toX, toY)
	if !isValidMove {
		return false
	}

	g.RemovePieceAt(toX, toY)

	piece.x = toX
	piece.y = toY

	g.turn++

	g.moves = append(g.moves, Move{piece.color, piece.type_, fromX, fromY, toX, toY, moveType})

	return true
}

func (g *Game) IsMoveValid(piece Piece, toX int, toY int) (bool, int) {
	// out of bounds
	if toX < 0 || toX > 7 || toY < 0 || toY > 7 {
		return false, constants.NonSpecialMove
	}

	// cant move to same place
	if piece.x == toX && piece.y == toY {
		return false, constants.NonSpecialMove
	}

	isValidMove := false
	moveType := constants.NonSpecialMove

	switch piece.type_ {
	case constants.Pawn:
		isValidMove, moveType = g.IsPawnMoveValid(piece, toX, toY)
	case constants.Rook:
		isValidMove, moveType = g.IsRookMoveValid(piece, toX, toY), constants.NonSpecialMove
	case constants.Knight:
		isValidMove, moveType = g.IsKnightMoveValid(piece, toX, toY), constants.NonSpecialMove
	case constants.Bishop:
		isValidMove, moveType = g.IsBishopMoveValid(piece, toX, toY), constants.NonSpecialMove
	case constants.Queen:
		isValidMove, moveType = g.IsQueenMoveValid(piece, toX, toY), constants.NonSpecialMove
	case constants.King:
		return g.IsKingMoveValid(piece, toX, toY)
	}

	if !isValidMove {
		return false, constants.NonSpecialMove
	}

	// check if move puts own king in check
	// temporarily move piece
	pieceRef := g.GetPieceAt(piece.x, piece.y)
	pieceRef.x = toX
	pieceRef.y = toY

	// check if king is in check
	if g.IsInCheck(piece.color) {
		isValidMove = false
		moveType = constants.NonSpecialMove
	}

	// move piece back
	pieceRef.x = piece.x
	pieceRef.y = piece.y

	return isValidMove, moveType
}

func (g *Game) IsPawnMoveValid(piece Piece, toX int, toY int) (bool, int) {
	direction := 1
	if piece.color == constants.Black {
		direction = -1
	}

	moveType := constants.NonSpecialMove

	// if promotion move
	if g.IsPromotionMove(piece, toX, toY) {
		moveType = constants.Promotion
	}

	isDestinationEmpty := g.GetPieceAt(toX, toY) == nil

	// if starting position, can move 2 squares
	if (piece.color == constants.White && piece.y == 1) || (piece.color == constants.Black && piece.y == 6) {
		if piece.x == toX && piece.y+direction*2 == toY {
			if isDestinationEmpty {
				return true, moveType
			}
		}
	}

	// can move 1 square
	if piece.x == toX && piece.y+direction == toY {
		if isDestinationEmpty {
			return true, moveType
		}
	}

	// can move 1 square diagonally
	xDiff := abs(piece.x - toX)

	isDestinationEnPassant := g.IsDestinationEnPassant(toX)

	if xDiff == 1 && piece.y+direction == toY {
		if !isDestinationEmpty {
			return true, moveType
		}
		if isDestinationEnPassant {
			return true, constants.EnPassant
		}
	}

	return false, moveType
}

func (g *Game) IsPromotionMove(piece Piece, toX int, toY int) bool {
	if piece.color == constants.White && toY == 7 {
		return true
	}

	if piece.color == constants.Black && toY == 0 {
		return true
	}

	return false
}

func (g *Game) IsDestinationEnPassant(toX int) bool {
	if len(g.moves) == 0 {
		return false
	}

	lastMove := g.moves[len(g.moves)-1]

	// has to be a pawn
	if lastMove.t != constants.Pawn {
		return false
	}

	// has to be a double move
	yDiff := abs(lastMove.fromY - lastMove.toY)
	if yDiff != 2 {
		return false
	}

	// has to be on the same file
	if lastMove.toX != toX {
		return false
	}

	return true
}

func (g *Game) IsRookMoveValid(piece Piece, toX int, toY int) bool {
	// can move horizontally or vertically but not both
	if piece.x != toX && piece.y != toY {
		return false
	}

	// can'type_ move through pieces
	if piece.x == toX {
		yDiff := abs(piece.y - toY)
		for i := 1; i < yDiff; i++ {
			if g.GetPieceAt(piece.x, piece.y+i) != nil {
				return false
			}
		}
	} else {
		xDiff := abs(piece.x - toX)
		for i := 1; i < xDiff; i++ {
			if g.GetPieceAt(piece.x+i, piece.y) != nil {
				return false
			}
		}
	}

	return true
}

func (g *Game) IsKnightMoveValid(piece Piece, toX int, toY int) bool {
	// can move 2 squares horizontally and 1 square vertically or vice versa
	xDiff := abs(piece.x - toX)
	yDiff := abs(piece.y - toY)

	if (xDiff == 2 && yDiff == 1) || (xDiff == 1 && yDiff == 2) {
		return true
	}

	return false
}

func (g *Game) IsBishopMoveValid(piece Piece, toX int, toY int) bool {
	xDiff := abs(piece.x - toX)
	yDiff := abs(piece.y - toY)

	// check that the difference in x and y is the same (diagonal)
	if xDiff != yDiff {
		return false
	}

	// can'type_ move through pieces
	for i := 1; i < xDiff; i++ {
		if g.GetPieceAt(piece.x+i, piece.y+i) != nil {
			return false
		}
	}

	return true
}

func (g *Game) IsQueenMoveValid(piece Piece, toX int, toY int) bool {
	// can move horizontally, vertically, or diagonally
	if g.IsRookMoveValid(piece, toX, toY) || g.IsBishopMoveValid(piece, toX, toY) {
		return true
	}

	return false
}

func (g *Game) IsKingMoveValid(piece Piece, toX int, toY int) (bool, int) {
	// can move 1 square in any direction
	xDiff := abs(piece.x - toX)
	yDiff := abs(piece.y - toY)

	if xDiff <= 1 && yDiff <= 1 {
		return true, constants.NonSpecialMove
	}

	// cant move directly next to enemy king
	enemyKing := g.GetKing(constants.GetOppositeColor(piece.color))
	if enemyKing != nil {
		enemyKingXDiff := abs(enemyKing.x - toX)
		enemyKingYDiff := abs(enemyKing.y - toY)
		if enemyKingXDiff <= 1 && enemyKingYDiff <= 1 {
			return false, constants.NonSpecialMove
		}
	}

	// can'type_ castle if king would be in check
	if g.IsInCheckAt(toX, toY) {
		return false, constants.NonSpecialMove
	}

	// can castle
	if g.IsKingSideCastle(piece, toX, toY) || g.IsQueenSideCastle(piece, toX, toY) {
		return true, constants.Castling
	}

	return false, constants.NonSpecialMove
}

func (g *Game) IsKingSideCastle(piece Piece, toX int, toY int) bool {
	// can'type_ castle if king has moved
	if piece.hasMoved {
		return false
	}

	// can'type_ castle if rook has moved
	rook := g.GetPieceAt(7, piece.y)
	if rook == nil || rook.hasMoved {
		return false
	}

	// can'type_ castle if there are pieces in the way
	if g.GetPieceAt(5, piece.y) != nil || g.GetPieceAt(6, piece.y) != nil {
		return false
	}

	// can'type_ castle if king is in check
	if g.IsInCheck(piece.color) {
		return false
	}

	// can'type_ castle if king would be in check
	if g.IsInCheckAt(6, piece.y) {
		return false
	}

	return true
}

func (g *Game) IsQueenSideCastle(piece Piece, toX int, toY int) bool {
	// can'type_ castle if king has moved
	if piece.hasMoved {
		return false
	}

	// can'type_ castle if rook has moved
	rook := g.GetPieceAt(0, piece.y)
	if rook == nil || rook.hasMoved {
		return false
	}

	// can'type_ castle if there are pieces in the way
	if g.GetPieceAt(1, piece.y) != nil || g.GetPieceAt(2, piece.y) != nil || g.GetPieceAt(3, piece.y) != nil {
		return false
	}

	// can'type_ castle if king is in check
	if g.IsInCheck(piece.color) {
		return false
	}

	// can'type_ castle if king would be in check
	if g.IsInCheckAt(2, piece.y) {
		return false
	}

	return true
}

func (g *Game) IsInCheck(color int) bool {
	king := g.GetKing(color)

	return g.IsInCheckAt(king.x, king.y)
}

func (g *Game) IsInCheckmate(color int) bool {
	king := g.GetKing(color)

	// check if king is in check
	if !g.IsInCheck(color) {
		return false
	}

	// check if king can move
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			targetX := king.x + x
			targetY := king.y + y

			isValidMove, _ := g.IsMoveValid(*king, targetX, targetY)
			if isValidMove {
				return false
			}
		}
	}

	// get all enemy pieces that check the king
	checkingPieces := make([]Piece, 0)
	for _, piece := range g.pieces {
		opponentColor := constants.GetOppositeColor(color)
		if piece.color == opponentColor && piece.type_ != constants.King {
			// if piece can move onto the king
			pieceIsCheckingKing, _ := g.IsMoveValid(piece, king.x, king.y)
			if pieceIsCheckingKing {
				checkingPieces = append(checkingPieces, piece)
			}
		}
	}

	// if there is more than one piece checking the king, then the king is in checkmate
	if len(checkingPieces) > 1 {
		return true
	}

	checkingPiece := checkingPieces[0]

	// check if the piece checking the king can be captured
	for _, piece := range g.pieces {
		if piece.color == color {
			isValidMove, _ := g.IsMoveValid(piece, checkingPiece.x, checkingPiece.y)
			if isValidMove {
				return false
			}
		}
	}

	// knights can't be blocked
	if checkingPiece.type_ == constants.Knight {
		return true
	}

	// check if it is in the same row
	if checkingPiece.x == king.x {
		// check if any player piece can be placed between the king and the checking piece
		for y := min(checkingPiece.y, king.y) + 1; y < max(checkingPiece.y, king.y); y++ {
			for _, piece := range g.pieces {
				if piece.color == color && piece.type_ != constants.King {
					isValidMove, _ := g.IsMoveValid(piece, checkingPiece.x, y)
					if isValidMove {
						return false
					}
				}
			}
		}

		return true
	}

	// check if it is in the same column
	if checkingPiece.y == king.y {
		// check if any player piece can be placed between the king and the checking piece
		for x := min(checkingPiece.x, king.x) + 1; x < max(checkingPiece.x, king.x); x++ {
			for _, piece := range g.pieces {
				if piece.color == color && piece.type_ != constants.King {
					isValidMove, _ := g.IsMoveValid(piece, x, checkingPiece.y)
					if isValidMove {
						return false
					}
				}
			}
		}

		return true
	}

	// check if it is in the same diagonal
	if abs(checkingPiece.x-king.x) == abs(checkingPiece.y-king.y) {
		startX := min(checkingPiece.x, king.x) + 1
		startY := min(checkingPiece.y, king.y) + 1

		diff := abs(checkingPiece.x-king.x) - 1

		for i := 0; i < diff; i++ {
			for _, piece := range g.pieces {
				if piece.color == color && piece.type_ != constants.King {
					isValidMove, _ := g.IsMoveValid(piece, startX+i, startY+i)
					if isValidMove {
						return false
					}
				}
			}
		}

		return true
	}

	return true
}

func (g *Game) IsInStalemate(color int) bool {
	king := g.GetKing(color)

	// check if king is in check
	if g.IsInCheck(color) {
		return false
	}

	// check if king can move
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			targetX := king.x + x
			targetY := king.y + y

			isValidMove, _ := g.IsMoveValid(*king, targetX, targetY)
			if isValidMove {
				return false
			}
		}
	}

	// check if any player piece can move
	for _, piece := range g.pieces {
		if piece.color == color && piece.type_ != constants.King {
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					isValidMove, _ := g.IsMoveValid(piece, x, y)
					if isValidMove {
						return false
					}
				}
			}
		}
	}

	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (g *Game) GetKing(color int) *Piece {
	for _, piece := range g.pieces {
		if piece.color == color && piece.type_ == constants.King {
			return &piece
		}
	}

	return nil
}

func (g *Game) IsInCheckAt(x int, y int) bool {
	for _, piece := range g.pieces {
		if piece.color != g.ActiveColor() && piece.type_ != constants.King {
			isValidMove, _ := g.IsMoveValid(piece, x, y)
			if isValidMove {
				return true
			}
		}
	}

	return false
}

func (g *Game) GetPieceAt(x int, y int) *Piece {
	for i, piece := range g.pieces {
		if piece.x == x && piece.y == y {
			return &g.pieces[i]
		}
	}

	return nil
}

func (g *Game) RemovePieceAt(x int, y int) {
	for i, piece := range g.pieces {
		if piece.x == x && piece.y == y {
			g.pieces[i] = g.pieces[len(g.pieces)-1]
			g.pieces = g.pieces[:len(g.pieces)-1]
			return
		}
	}
}

type Player struct {
	name         string
	token        string
	connectionId string
	color        int
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) Token() string {
	return p.token
}

func (p *Player) Color() int {
	return p.color
}

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}