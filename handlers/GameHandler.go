package handlers

import (
	"encoding/json"
	"github.com/racccoooon/chess-be/constants"
	"github.com/racccoooon/chess-be/game"
	"net/http"
	"regexp"
	"strconv"
)

type GameHandler struct {
	manager *game.Manager
}

func NewGameHandler(manager *game.Manager) *GameHandler {
	return &GameHandler{manager: manager}
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/games/":
		h.newGame(w, r)
		return
	}

	var token string
	var gameId string
	var fromX int
	var fromY int

	// read token from header
	if tokenHeader := r.Header.Get("Authorization"); tokenHeader != "" {
		if len(tokenHeader) < 7 || tokenHeader[:7] != "Bearer " {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// token head is "Bearer " + token
		token = tokenHeader[7:]
	}

	switch {
	case r.Method == http.MethodGet && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/validmoves/([0-7])/([0-7])$", &gameId, &fromX, &fromY):
		h.getValidMoves(w, r, token, game.Id(gameId), fromX, fromY)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

var cachedRegex = map[string]*regexp.Regexp{}

func mustCompileCached(pattern string) *regexp.Regexp {
	if regex, ok := cachedRegex[pattern]; ok {
		return regex
	}
	regex := regexp.MustCompile(pattern)
	cachedRegex[pattern] = regex
	return regex
}

func match(path, pattern string, routeParams ...interface{}) bool {
	regex := mustCompileCached(pattern)
	matches := regex.FindStringSubmatch(path)

	if len(matches) <= 0 {
		return false
	}

	for i, matchValue := range matches[1:] {
		switch param := routeParams[i].(type) {
		case *string:
			*param = matchValue

		case *int:
			numberValue, err := strconv.Atoi(matchValue)
			if err != nil {
				return false
			}
			*param = numberValue

		default:
			panic("routeParams must be *string or *int")
		}
	}

	return true
}

type newGameRequest struct {
	Color          string          `json:"color"` // white or black or randomColor
	StartingPieces []StartingPiece `json:"startingPieces"`
	StartingColor  string          `json:"startingColor"`
}

type StartingPiece struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

type newGameResponse struct {
	GameId string `json:"gameId"`
}

func (h *GameHandler) newGame(w http.ResponseWriter, r *http.Request) {
	var request newGameRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	startingPieces := make([]game.Piece, len(request.StartingPieces))
	for i, startingPiece := range request.StartingPieces {
		startingPieces[i] = game.NewPiece(
			constants.ColorFromString(startingPiece.Color),
			constants.TypeFromString(startingPiece.Type),
			startingPiece.X,
			startingPiece.Y)
	}

	game := h.manager.NewGame(constants.ColorFromString(request.Color), startingPieces, constants.ColorFromString(request.StartingColor))

	response := newGameResponse{
		GameId: string(game.Id()),
	}

	responseMessage, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(responseMessage)
	w.WriteHeader(http.StatusCreated)
}

type validMovesResponse struct {
	ValidMoves []validMoveResponseItem `json:"validMoves"`
}

type validMoveResponseItem struct {
	ToX int `json:"toX"`
	ToY int `json:"toY"`
}

func (h *GameHandler) getValidMoves(w http.ResponseWriter, r *http.Request, token string, gameId game.Id, fromX, fromY int) {
	game := h.manager.GetGame(gameId)
	if game == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if game.GetPlayerByToken(token) == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	validMoves := game.GetValidMoves(fromX, fromY)

	response := validMovesResponse{
		ValidMoves: make([]validMoveResponseItem, len(validMoves)),
	}

	for i, validMove := range validMoves {
		response.ValidMoves[i] = validMoveResponseItem{
			ToX: validMove.ToX(),
			ToY: validMove.ToY(),
		}
	}

	responseMessage, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(responseMessage)
}
