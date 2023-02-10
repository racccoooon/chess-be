package main

import (
	"github.com/racccoooon/chess-be/game"
	"github.com/racccoooon/chess-be/handlers"
	"github.com/racccoooon/chess-be/hubs"
	"github.com/racccoooon/chess-be/middlewares"
	"net/http"
)

func main() {
	router := http.NewServeMux()

	gameManager := game.NewGameManager()

	hubs.SetupGameHub(gameManager, router)

	router.Handle("/api/games/", handlers.NewGameHandler(gameManager))
	router.Handle("/api/health", handlers.NewHealthHandler())

	corsMiddleware := &middlewares.CorsMiddleware{Handler: router}

	err := http.ListenAndServe(":8080", corsMiddleware)
	if err != nil {
		panic(err)
	}
}
