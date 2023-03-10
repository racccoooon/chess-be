package hubs

import (
	"context"
	"github.com/go-kit/log"
	"github.com/philippseith/signalr"
	"github.com/racccoooon/chess-be/constants"
	"github.com/racccoooon/chess-be/game"
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
	Board           []BoardItemResponse `json:"board"`
	InitialBoard    []BoardItemResponse `json:"initialBoard"`
	Moves           []MoveItemResponse  `json:"moves"`
	ActiveColor     string              `json:"activeColor"`
	PlayerColor     string              `json:"playerColor"`
	WhitePlayerName string              `json:"whitePlayerName"`
	BlackPlayerName string              `json:"blackPlayerName"`
	StartingColor   string              `json:"startingColor"`
}

type BoardItemResponse struct {
	Color    string      `json:"color"`
	Type     string      `json:"type"`
	Position PositionDto `json:"position"`
}

type MoveItemResponse struct {
	From          PositionDto `json:"from"`
	To            PositionDto `json:"to"`
	Color         string      `json:"color"`
	Type          string      `json:"type"`
	Kind          string      `json:"kind"`
	Status        string      `json:"status"`
	Captures      bool        `json:"captures"`
	PromoteToType *string     `json:"promoteToType"`
}

type PositionDto struct {
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
		Board:           []BoardItemResponse{},
		Moves:           []MoveItemResponse{},
		ActiveColor:     constants.ColorAsString(game.ActiveColor()),
		PlayerColor:     constants.ColorAsString(player.Color()),
		WhitePlayerName: game.OpponentName(constants.Black),
		BlackPlayerName: game.OpponentName(constants.White),
		StartingColor:   constants.ColorAsString(game.StartingColor()),
	}

	for _, piece := range game.Pieces() {
		joinResponse.Board = append(joinResponse.Board, BoardItemResponse{
			Color: constants.ColorAsString(piece.Color()),
			Type:  constants.TypeAsString(piece.Type()),
			Position: PositionDto{
				X: piece.X(),
				Y: piece.Y(),
			},
		})
	}

	for _, piece := range game.InitialPieces() {
		joinResponse.InitialBoard = append(joinResponse.InitialBoard, BoardItemResponse{
			Color: constants.ColorAsString(piece.Color()),
			Type:  constants.TypeAsString(piece.Type()),
			Position: PositionDto{
				X: piece.X(),
				Y: piece.Y(),
			},
		})
	}

	for _, move := range game.History() {
		joinResponse.Moves = append(joinResponse.Moves, moveAsMoveItem(move))
	}

	h.Clients().Caller().Send("gameJoined", joinResponse)

	h.Groups().AddToGroup("game-"+request.GameId, h.ConnectionID())

	if game.PlayerCount() == 2 {
		gameStartedResponse := GameStartedResponse{
			WhitePlayerName: game.OpponentName(constants.Black),
			BlackPlayerName: game.OpponentName(constants.White),
		}

		h.Clients().Group("game-"+request.GameId).Send("gameStarted", gameStartedResponse)
		h.Clients().Group("spectators-"+request.GameId).Send("gameStarted", gameStartedResponse)
	}
}

func moveAsMoveItem(move game.Move) MoveItemResponse {
	t := move.Type()
	if move.Kind() == constants.Promotion {
		t = constants.Pawn
	}

	var promotionType *string = nil
	if move.Kind() == constants.Promotion {
		temp := constants.TypeAsString(move.Type())
		promotionType = &temp
	}

	return MoveItemResponse{
		From: PositionDto{
			X: move.FromX(),
			Y: move.FromY(),
		},
		To: PositionDto{
			X: move.ToX(),
			Y: move.ToY(),
		},
		Color:         constants.ColorAsString(move.Color()),
		Type:          constants.TypeAsString(t),
		Kind:          constants.MoveKindAsString(move.Kind()),
		Status:        constants.StatusAsString(move.Status()),
		Captures:      move.Captures(),
		PromoteToType: promotionType,
	}
}

type JoinSpectatorRequest struct {
	GameId string `json:"gameId"`
}

func (h *GameHub) JoinSpectator(request JoinSpectatorRequest) {
	manager := h.Context().Value("manager").(*game.Manager)

	game := manager.GetGame(game.Id(request.GameId))
	if game == nil {
		h.Clients().Caller().Send("gameNotFound")
		return
	}

	joinResponse := JoinGameResponse{
		Board:           []BoardItemResponse{},
		Moves:           []MoveItemResponse{},
		ActiveColor:     constants.ColorAsString(game.ActiveColor()),
		PlayerColor:     "None",
		WhitePlayerName: game.OpponentName(constants.Black),
		BlackPlayerName: game.OpponentName(constants.White),
		StartingColor:   constants.ColorAsString(game.StartingColor()),
	}

	for _, piece := range game.Pieces() {
		joinResponse.Board = append(joinResponse.Board, BoardItemResponse{
			Color: constants.ColorAsString(piece.Color()),
			Type:  constants.TypeAsString(piece.Type()),
			Position: PositionDto{
				X: piece.X(),
				Y: piece.Y(),
			},
		})
	}

	for _, piece := range game.InitialPieces() {
		joinResponse.InitialBoard = append(joinResponse.InitialBoard, BoardItemResponse{
			Color: constants.ColorAsString(piece.Color()),
			Type:  constants.TypeAsString(piece.Type()),
			Position: PositionDto{
				X: piece.X(),
				Y: piece.Y(),
			},
		})
	}

	for _, move := range game.History() {
		joinResponse.Moves = append(joinResponse.Moves, moveAsMoveItem(move))
	}

	h.Clients().Caller().Send("gameJoined", joinResponse)

	h.Groups().AddToGroup("spectators-"+request.GameId, h.ConnectionID())
}

func (h *GameHub) LeaveSpectator(request JoinSpectatorRequest) {
	h.Groups().RemoveFromGroup("spectators-"+request.GameId, h.ConnectionID())
}

func (h *GameHub) gameNotFound() {
	h.Clients().Caller().Send("gameNotFound")
}

type MoveRequest struct {
	GameId        string      `json:"gameId"`
	From          PositionDto `json:"from"`
	To            PositionDto `json:"to"`
	PromoteToType *string     `json:"promoteToType"`
}

func (h *GameHub) Move(request MoveRequest) {
	manager := h.Context().Value("manager").(*game.Manager)

	game := manager.GetGame(game.Id(request.GameId))
	if game == nil {
		h.gameNotFound()
		return
	}

	player := game.GetPlayerByConnectionId(h.ConnectionID())
	if player == nil {
		h.Clients().Caller().Send("playerNotFound")
	}

	if game.ActiveColor() != player.Color() {
		return
	}

	// check if move is valid

	if !game.Move(request.From.X, request.From.Y, request.To.X, request.To.Y, request.PromoteToType) {
		h.Clients().Caller().Send("invalidMove")
	}

	lastMove := game.LastMove()

	moveItemResponse := moveAsMoveItem(*lastMove)
	h.Clients().Group("game-"+request.GameId).Send("move", moveItemResponse)
	h.Clients().Group("spectators-"+request.GameId).Send("move", moveItemResponse)
}

type ChangeNameRequest struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type ChangeNameResponse struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *GameHub) ChangeName(request ChangeNameRequest) {
	manager := h.Context().Value("manager").(*game.Manager)

	playerGames := manager.GetGamesForPlayer(request.Token)

	for _, playerGame := range playerGames {
		h.Clients().Group("game-"+string(playerGame.Id())).Send("playerNameChanged", ChangeNameResponse{
			Name:  request.Name,
			Color: constants.ColorAsString(playerGame.Color()),
		})

		h.Clients().Group("spectators-"+string(playerGame.Id())).Send("playerNameChanged", ChangeNameResponse{
			Name:  request.Name,
			Color: constants.ColorAsString(playerGame.Color()),
		})
	}
}
