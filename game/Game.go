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
	color int
	t     int
	x     int
	y     int
}

func (Piece *Piece) Color() int {
	return Piece.color
}

func (Piece *Piece) Type() int {
	return Piece.t
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
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Rook, x: 0, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Knight, x: 1, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Bishop, x: 2, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Queen, x: 3, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.King, x: 4, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Bishop, x: 5, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Knight, x: 6, y: 0})
	g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Rook, x: 7, y: 0})

	for i := 0; i < 8; i++ {
		g.pieces = append(g.pieces, Piece{color: constants.White, t: constants.Pawn, x: i, y: 1})
	}

	// black
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Rook, x: 0, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Knight, x: 1, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Bishop, x: 2, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Queen, x: 3, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.King, x: 4, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Bishop, x: 5, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Knight, x: 6, y: 7})
	g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Rook, x: 7, y: 7})

	for i := 0; i < 8; i++ {
		g.pieces = append(g.pieces, Piece{color: constants.Black, t: constants.Pawn, x: i, y: 6})
	}
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
