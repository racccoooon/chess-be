package main

import (
	"net/http"
	"regexp"
	"strconv"
)

type gameHandler struct{}

func (h *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var gameId string
	var token string

	setDefaultHeaders(w, r)

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/games":
		h.newGame(w, r)
		return

	case r.Method == http.MethodGet && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/join$", &gameId):
		h.joinGame(w, r, gameId)
		return

	case r.Method == http.MethodPut && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/players/([a-zA-Z0-9-]+)", &gameId, &token):
		h.updatePlayer(w, r, gameId, token)
		return

	case r.Method == http.MethodGet && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/move$", &gameId):
		h.handleMove(w, r, gameId)
		return

	case r.Method == http.MethodPost && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/players/([a-zA-Z0-9-]+)/forfeit", &gameId, &token):
		h.forfeit(w, r, gameId, token)
		return
	}
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

func (h *gameHandler) newGame(w http.ResponseWriter, r *http.Request) {

}

func (h *gameHandler) handleMove(w http.ResponseWriter, r *http.Request, gameId string) {
	// write hello world
	w.Write([]byte(gameId))
}

func (h *gameHandler) joinGame(w http.ResponseWriter, r *http.Request, gameId string) {

}

func (h *gameHandler) updatePlayer(w http.ResponseWriter, r *http.Request, id string, token string) {

}

func (h *gameHandler) forfeit(w http.ResponseWriter, r *http.Request, id string, token string) {

}

type healthHandler struct{}

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w, r)
	w.WriteHeader(http.StatusOK)
}

func setDefaultHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

type Game struct {
	Id string
}

var games = make(map[string]Game)

func main() {
	router := http.NewServeMux()

	router.Handle("/api/games/", &gameHandler{})
	router.Handle("/api/health", &healthHandler{})

	http.ListenAndServe(":8080", router)
}
