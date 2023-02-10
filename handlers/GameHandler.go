package handlers

import (
	"encoding/json"
	"github.com/racccoooon/chess-be/constants"
	"github.com/racccoooon/chess-be/game"
	"net/http"
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

	w.WriteHeader(http.StatusNotFound)
}

type newGameRequest struct {
	Color string `json:"color"` // white or black or randomColor
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

	game := h.manager.NewGame(constants.ColorFromString(request.Color))

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
