package hubs

import (
	"chess-be/constants"
	"chess-be/game"
	"context"
	"github.com/go-kit/log"
	"github.com/philippseith/signalr"
	"net/http"
	"os"
	"time"
)

type GameHub struct {
	signalr.Hub
}

func SetupGameHub(manager *game.Manager, router *http.ServeMux) {
	hub := &GameHub{}

	context := context.WithValue(context.Background(), "manager", manager)

	server, err := signalr.NewServer(context,
		signalr.SimpleHubFactory(hub),
		signalr.HTTPTransports("ServerSentEvents"),
		signalr.KeepAliveInterval(2*time.Second),
		signalr.TimeoutInterval(10*time.Second),
		signalr.Logger(log.NewLogfmtLogger(os.Stdout), true),
		signalr.EnableDetailedErrors(true))

	if err != nil {
		panic(err)
	}

	server.MapHTTP(signalr.WithHTTPServeMux(router), "/gameHub")
}

type JoinGameRequest struct {
	GameId     string `json:"gameId"`
	PlayerName string `json:"playerName"`
	Token      string `json:"token"`
}

type JoinGameResponse struct {
	Board        []BoardItemResponse `json:"board"`
	Moves        []MoveItemResponse  `json:"moves"`
	ActiveColor  string              `json:"activeColor"`
	PlayerColor  string              `json:"playerColor"`
	OpponentName string              `json:"opponentName"`
}

type BoardItemResponse struct {
	Color    string           `json:"color"`
	Type     string           `json:"type"`
	Position PositionResponse `json:"position"`
}

type MoveItemResponse struct {
	From  PositionResponse `json:"from"`
	To    PositionResponse `json:"to"`
	Color string           `json:"color"`
	Type  string           `json:"type"`
}

type PositionResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type GameStartedResponse struct {
	WhitePlayerName string `json:"whitePlayerName"`
	BlackPlayerName string `json:"blackPlayerName"`
}

func (h *GameHub) JoinGame(request JoinGameRequest) {
	manager := h.Context().Value("manager").(*game.Manager)

	game := manager.GetGame(game.Id(request.GameId))
	if game == nil {
		h.gameNotFound()
		return
	}

	player := game.GetPlayerByToken(request.Token)
	if game.PlayerCount() == 2 {
		if player == nil {
			h.Clients().Caller().Send("gameFull")
			return
		}
	}

	if player != nil {
		player = game.RejoinPlayer(request.Token, h.ConnectionID())
	} else {
		player = game.AddPlayer(request.PlayerName, request.Token, h.ConnectionID())
	}

	joinResponse := JoinGameResponse{
		Board:        []BoardItemResponse{},
		Moves:        []MoveItemResponse{},
		ActiveColor:  constants.ColorAsString(game.ActiveColor()),
		PlayerColor:  constants.ColorAsString(player.Color()),
		OpponentName: game.OpponentName(player.Color()),
	}

	for _, piece := range game.Pieces() {
		joinResponse.Board = append(joinResponse.Board, BoardItemResponse{
			Color: constants.ColorAsString(piece.Color()),
			Type:  constants.TypeAsString(piece.Type()),
			Position: PositionResponse{
				X: piece.X(),
				Y: piece.Y(),
			},
		})
	}

	for _, move := range game.History() {
		joinResponse.Moves = append(joinResponse.Moves, MoveItemResponse{
			From: PositionResponse{
				X: move.FromX(),
				Y: move.FromY(),
			},
			To: PositionResponse{
				X: move.ToX(),
				Y: move.ToY(),
			},
			Color: constants.ColorAsString(move.Color()),
			Type:  constants.TypeAsString(move.Type()),
		})
	}

	h.Clients().Caller().Send("gameJoined", joinResponse)

	h.Groups().AddToGroup("game-"+request.GameId, h.ConnectionID())

	if game.PlayerCount() == 2 {
		gameStartedResponse := GameStartedResponse{
			WhitePlayerName: game.OpponentName(constants.Black),
			BlackPlayerName: game.OpponentName(constants.White),
		}

		h.Clients().Group("game-"+request.GameId).Send("gameStarted", gameStartedResponse)
	}
}

func (h *GameHub) JoinSpectator(request JoinGameRequest) {
	manager := h.Context().Value("manager").(*game.Manager)

	game := manager.GetGame(game.Id(request.GameId))
	if game == nil {
		h.Clients().Caller().Send("gameNotFound")
		return
	}

	h.Groups().AddToGroup("spectators-"+request.GameId, h.ConnectionID())
}

func (h *GameHub) LeaveSpectator(request JoinGameRequest) {
	h.Groups().RemoveFromGroup("spectators-"+request.GameId, h.ConnectionID())
}

func (h *GameHub) gameNotFound() {
	h.Clients().Caller().Send("gameNotFound")
}

/*type MoveRequest struct {
	GameId string `json:"gameId"`
	FromX  int    `json:"fromX"`
	FromY  int    `json:"fromY"`
	ToX    int    `json:"toX"`
	ToY    int    `json:"toY"`
}

func (h *GameHub) Move(request MoveRequest) {
	game := h.manager.GetGame(game.Id(request.GameId))
	if game == nil {
		h.gameNotFound()
		return
	}

	player := game.GetPlayerByConnectionId(h.ConnectionID())
	if player == nil {
		h.Clients().Caller().Send("notYourTurn", "You are not allowed to move")
	}
}*/
