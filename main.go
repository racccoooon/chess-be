package main

import (
	"context"
	"encoding/json"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/philippseith/signalr"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type gameHandler struct {
	signalrServer *GameHub
}

func (h *gameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var gameId string
	var token string

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/games/":
		h.newGame(w, r)
		return

	case r.Method == http.MethodPost && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/join/$", &gameId):
		h.joinGame(w, r, gameId)
		return
	}

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
	case r.Method == http.MethodPut && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/players/", &gameId):
		h.updatePlayer(w, r, gameId, token)
		return

	case r.Method == http.MethodPost && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/moves/$", &gameId):
		h.handleMove(w, r, gameId, token)
		return

	case r.Method == http.MethodPost && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/promote/$", &gameId):
		h.promote(w, r, gameId, token)
		return

	case r.Method == http.MethodPost && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/forfeit/", &gameId):
		h.forfeit(w, r, gameId, token)
		return

	case r.Method == http.MethodGet && match(r.URL.Path, "^/api/games/([a-zA-Z0-9-]+)/", &gameId):
		h.getGame(w, r, gameId, token)
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

type newGameResponse struct {
	GameId     string `json:"gameId"`
	Token      string `json:"token"`
	PlayerName string `json:"playerName"`
}

func (h *gameHandler) newGame(w http.ResponseWriter, r *http.Request) {
	game := &Game{
		Id: uuid.New().String(),
		Players: []Player{
			{
				Token: uuid.New().String(),
				Name:  "Player 1",
				Color: white,
			},
		},
		Turn: 0,
	}

	game.SetupBoard()

	games[game.Id] = game

	response := newGameResponse{
		GameId:     game.Id,
		Token:      game.Players[0].Token,
		PlayerName: game.Players[0].Name,
	}

	responseMessage, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(responseMessage)
	w.WriteHeader(http.StatusCreated)
}

type joinGameResponse struct {
	Token        string `json:"token"`
	PlayerName   string `json:"playerName"`
	OpponentName string `json:"opponentName"`
}

func (h *gameHandler) joinGame(w http.ResponseWriter, r *http.Request, gameId string) {
	game, ok := games[gameId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if len(game.Players) >= 2 {
		w.WriteHeader(http.StatusConflict)
		return
	}

	player := Player{
		Token: uuid.New().String(),
		Name:  "Player 2",
		Color: black,
	}

	game.Players = append(game.Players, player)
	games[gameId] = game

	response := joinGameResponse{
		Token:        player.Token,
		PlayerName:   player.Name,
		OpponentName: game.Players[0].Name,
	}

	responseMessage, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(responseMessage)
	w.WriteHeader(http.StatusCreated)
}

type updatePlayerRequest struct {
	PlayerName string `json:"playerName"`
}

type updatePlayerEvent struct {
	PlayerName string `json:"playerName"`
	Color      string `json:"color"`
}

func (h *gameHandler) updatePlayer(w http.ResponseWriter, r *http.Request, id string, token string) {
	game, ok := games[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var request updatePlayerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	player.Name = request.PlayerName

	w.WriteHeader(http.StatusNoContent)

	updatePlayerEvent := updatePlayerEvent{
		PlayerName: player.Name,
		Color:      colorAsString(player.Color),
	}

	h.signalrServer.Clients().Group(game.Id).Send("opponentUpdated", updatePlayerEvent)
}

type moveRequest struct {
	FromCell cellRequest `json:"fromCell"`
	ToCell   cellRequest `json:"toCell"`
}

type cellRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (h *gameHandler) handleMove(w http.ResponseWriter, r *http.Request, gameId string, token string) {
	game, ok := games[gameId]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var request moveRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isValidMove := game.isValidMove(player, request.FromCell, request.ToCell)

	if !isValidMove {
		w.WriteHeader(http.StatusConflict)
		return
	}

	game.movePiece(player, request.FromCell, request.ToCell)

	w.WriteHeader(http.StatusNoContent)

	h.signalrServer.Hub.Clients().Group(game.Id).Send("moveMade", request.FromCell.X, request.FromCell.Y, request.ToCell.X, request.ToCell.Y)
}

type forfeitEvent struct {
	Color string `json:"color"`
}

func (h *gameHandler) forfeit(w http.ResponseWriter, r *http.Request, id string, token string) {
	game, ok := games[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	forfeitEvent := forfeitEvent{
		Color: colorAsString(player.Color),
	}

	h.signalrServer.Clients().Group(game.Id).Send("gameForfeited", forfeitEvent)
}

type promoteRequest struct {
	Type string      `json:"type"`
	Cell cellRequest `json:"cell"`
}

type promoteEvent struct {
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

func (h *gameHandler) promote(w http.ResponseWriter, r *http.Request, id string, token string) {
	game, ok := games[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var request promoteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isValidPromotion := game.IsValidPromotion(player, request.Cell)
	if !isValidPromotion {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game.Promote(player, request.Type, request.Cell)

	w.WriteHeader(http.StatusNoContent)

	promoteEvent := promoteEvent{
		Type: request.Type,
		X:    request.Cell.X,
		Y:    request.Cell.Y,
	}

	h.signalrServer.Clients().Group(game.Id).Send("piecePromoted", promoteEvent)
}

type getGameResponse struct {
	Pieces []pieceResponse `json:"pieces"`
}

type pieceResponse struct {
	Type  string `json:"type"`
	Color string `json:"color"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

func (h *gameHandler) getGame(w http.ResponseWriter, r *http.Request, id string, token string) {
	game, ok := games[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	response := getGameResponse{
		Pieces: []pieceResponse{},
	}

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			piece := game.Cells[x][y]
			if piece.Type != empty {
				response.Pieces = append(response.Pieces, pieceResponse{
					Type:  typeAsString(piece.Type),
					Color: colorAsString(piece.Color),
					X:     x,
					Y:     y,
				})
			}
		}
	}

	responseMessage, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(responseMessage)

	w.WriteHeader(http.StatusOK)
}

func colorAsString(color int) string {
	switch color {
	case white:
		return "white"
	case black:
		return "black"
	}

	panic("invalid color")
}

func typeAsString(t int) string {
	switch t {
	case pawn:
		return "pawn"
	case rook:
		return "rook"
	case knight:
		return "knight"
	case bishop:
		return "bishop"
	case queen:
		return "queen"
	case king:
		return "king"
	}

	panic("invalid type")
}

func getPlayer(game *Game, token string) *Player {
	for i := range game.Players {
		if game.Players[i].Token == token {
			return &game.Players[i]
		}
	}

	return nil
}

type healthHandler struct{}

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type Game struct {
	Id              string
	Players         []Player
	Cells           [8][8]cell
	Turn            int
	LastInteraction time.Time
}

func (g *Game) SetupBoard() {
	g.SetupPawns()
	g.SetupRooks()
	g.SetupKnights()
	g.SetupBishops()
	g.SetupQueens()
	g.SetupKings()
}

func (g *Game) SetupPawns() {
	for i := 0; i < 8; i++ {
		g.Cells[i][1] = cell{
			X:     i,
			Y:     1,
			Type:  pawn,
			Color: white,
		}

		g.Cells[i][6] = cell{
			X:     i,
			Y:     6,
			Type:  pawn,
			Color: black,
		}
	}
}

func (g *Game) SetupRooks() {
	g.Cells[0][0] = cell{
		X:     0,
		Y:     0,
		Type:  rook,
		Color: white,
	}

	g.Cells[7][0] = cell{
		X:     7,
		Y:     0,
		Type:  rook,
		Color: white,
	}

	g.Cells[0][7] = cell{
		X:     0,
		Y:     7,
		Type:  rook,
		Color: black,
	}

	g.Cells[7][7] = cell{
		X:     7,
		Y:     7,
		Type:  rook,
		Color: black,
	}
}

func (g *Game) SetupKnights() {
	g.Cells[1][0] = cell{
		X:     1,
		Y:     0,
		Type:  knight,
		Color: white,
	}

	g.Cells[6][0] = cell{
		X:     6,
		Y:     0,
		Type:  knight,
		Color: white,
	}

	g.Cells[1][7] = cell{
		X:     1,
		Y:     7,
		Type:  knight,
		Color: black,
	}

	g.Cells[6][7] = cell{
		X:     6,
		Y:     7,
		Type:  knight,
		Color: black,
	}
}

func (g *Game) SetupBishops() {
	g.Cells[2][0] = cell{
		X:     2,
		Y:     0,
		Type:  bishop,
		Color: white,
	}

	g.Cells[5][0] = cell{
		X:     5,
		Y:     0,
		Type:  bishop,
		Color: white,
	}

	g.Cells[2][7] = cell{
		X:     2,
		Y:     7,
		Type:  bishop,
		Color: black,
	}

	g.Cells[5][7] = cell{
		X:     5,
		Y:     7,
		Type:  bishop,
		Color: black,
	}
}

func (g *Game) SetupQueens() {
	g.Cells[3][0] = cell{
		X:     3,
		Y:     0,
		Type:  queen,
		Color: white,
	}

	g.Cells[3][7] = cell{
		X:     3,
		Y:     7,
		Type:  queen,
		Color: black,
	}
}

func (g *Game) SetupKings() {
	g.Cells[4][0] = cell{
		X:     4,
		Y:     0,
		Type:  king,
		Color: white,
	}

	g.Cells[4][7] = cell{
		X:     4,
		Y:     7,
		Type:  king,
		Color: black,
	}
}

func (g *Game) isValidMove(player *Player, fromCell cellRequest, toCell cellRequest) bool {
	if g.Turn%2 != player.Color {
		return false
	}

	fromGamePiece := g.Cells[fromCell.X][fromCell.Y]
	if fromGamePiece.Color != player.Color {
		return false
	}

	toGamePiece := g.Cells[toCell.X][toCell.Y]

	switch fromGamePiece.Type {
	case pawn:
		return g.isValidPawnMove(fromGamePiece, toGamePiece)
	case rook:
		return g.isValidRookMove(fromGamePiece, toGamePiece)
	case knight:
		return g.isValidKnightMove(fromGamePiece, toGamePiece)
	case bishop:
		return g.isValidBishopMove(fromGamePiece, toGamePiece)
	case queen:
		return g.isValidQueenMove(fromGamePiece, toGamePiece)
	case king:
		return g.isValidKingMove(fromGamePiece, toGamePiece)
	}

	return false
}

func (g *Game) isValidPawnMove(from cell, to cell) bool {
	/*direction := 1
	if from.Color == black {
		direction = -1
	}*/

	return true
}

func (g *Game) isValidRookMove(piece cell, piece2 cell) bool {
	return true
}

func (g *Game) isValidKnightMove(piece cell, piece2 cell) bool {
	return true
}

func (g *Game) isValidBishopMove(piece cell, piece2 cell) bool {
	return true
}

func (g *Game) isValidQueenMove(piece cell, piece2 cell) bool {
	return true
}

func (g *Game) isValidKingMove(piece cell, piece2 cell) bool {
	return true
}

func (g *Game) IsValidPromotion(player *Player, request cellRequest) bool {
	gameCell := g.Cells[request.X][request.Y]

	if gameCell.Type != pawn {
		return false
	}

	if gameCell.Color == white && request.Y == 7 {
		return true
	}

	if gameCell.Color == black && request.Y == 0 {
		return true
	}

	return false
}

func (g *Game) Promote(player *Player, t string, request cellRequest) {
	gameCell := g.Cells[request.X][request.Y]

	switch t {
	case "rook":
		gameCell.Type = rook
		return
	case "knight":
		gameCell.Type = knight
		return
	case "bishop":
		gameCell.Type = bishop
		return
	case "queen":
		gameCell.Type = queen
		return
	default:
		panic("Invalid promotion type")
	}
}

func (g *Game) movePiece(player *Player, fromCell cellRequest, toCell cellRequest) {
	fromGamePiece := g.Cells[fromCell.X][fromCell.Y]
	toGamePiece := g.Cells[toCell.X][toCell.Y]

	g.Cells[toCell.X][toCell.Y] = fromGamePiece

	var isCastling = false
	if player.Color == white && fromGamePiece.Type == king && fromGamePiece.X == 4 && fromGamePiece.Y == 0 {
		if toGamePiece.Type == rook && toGamePiece.X == 7 && toGamePiece.Y == 0 {
			g.Cells[5][0] = g.Cells[7][0]
			g.Cells[7][0] = cell{}

			isCastling = true
		}

		if toGamePiece.Type == rook && toGamePiece.X == 0 && toGamePiece.Y == 0 {
			g.Cells[3][0] = g.Cells[0][0]
			g.Cells[0][0] = cell{}

			isCastling = true
		}
	}

	if player.Color == black && fromGamePiece.Type == king && fromGamePiece.X == 4 && fromGamePiece.Y == 7 {
		if toGamePiece.Type == rook && toGamePiece.X == 7 && toGamePiece.Y == 7 {
			g.Cells[5][7] = g.Cells[7][7]
			g.Cells[7][7] = cell{}

			isCastling = true
		}

		if toGamePiece.Type == rook && toGamePiece.X == 0 && toGamePiece.Y == 7 {
			g.Cells[3][7] = g.Cells[0][7]
			g.Cells[0][7] = cell{}

			isCastling = true
		}
	}

	if !isCastling {
		g.Cells[fromCell.X][fromCell.Y] = cell{}
	}

	g.Turn = (g.Turn + 1) % 2
}

type Player struct {
	Token string
	Name  string
	Color int
}

type cell struct {
	X     int
	Y     int
	Type  int
	Color int
}

const (
	empty = iota
	pawn
	rook
	knight
	bishop
	queen
	king
)

const (
	white = iota
	black
)

var games = make(map[string]*Game)

type CorsMiddleware struct {
	Handler http.Handler
}

var requestIDs = make(map[string]bool)

func (c *CorsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With, X-HTTP-Method-Override, X-Request-Id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// check if the request ID is known
	requestId := r.Header.Get("X-Request-Id")
	if requestId == "" {
		// if it's not, then it's not a duplicate request
		c.Handler.ServeHTTP(w, r)
		return
	}
	if _, ok := requestIDs[requestId]; ok {
		// if it is, then it's a duplicate request
		return
	}

	// otherwise, add it to the map
	requestIDs[requestId] = true

	// remove the request ID from the map after 5 minutes
	time.AfterFunc(5*time.Minute, func() {
		delete(requestIDs, requestId)
	})

	c.Handler.ServeHTTP(w, r)
}

type GameHub struct {
	signalr.Hub
}

func (g *GameHub) Join(gameId string, token string) {
	game := games[gameId]
	if game == nil {
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		return
	}

	g.Clients().Group(gameId).Send("gameJoined", player.Name)
	g.Hub.Groups().AddToGroup(gameId, g.ConnectionID())
}

func (g *GameHub) Leave(gameId string, token string) {
	game := games[gameId]
	if game == nil {
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		return
	}

	g.Hub.Groups().RemoveFromGroup(gameId, g.ConnectionID())
}

func (g *GameHub) UpdatePlayer(gameId string, token string, request updatePlayerRequest) {
	game, ok := games[gameId]
	if !ok {
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		return
	}

	player.Name = request.PlayerName

	updatePlayerEvent := updatePlayerEvent{
		PlayerName: player.Name,
		Color:      colorAsString(player.Color),
	}

	g.Hub.Clients().Group(game.Id).Send("opponentUpdated", updatePlayerEvent)
}

func (g *GameHub) MakeMove(gameId string, token string, request moveRequest) {
	game, ok := games[gameId]
	if !ok {
		return
	}

	player := getPlayer(game, token)

	if player == nil {
		return
	}

	isValidMove := game.isValidMove(player, request.FromCell, request.ToCell)

	if !isValidMove {
		return
	}

	game.movePiece(player, request.FromCell, request.ToCell)

	g.Hub.Clients().Group(game.Id).Send("moveMade", request.FromCell.X, request.FromCell.Y, request.ToCell.X, request.ToCell.Y)
}

func main() {

	hub := &GameHub{}

	server, err := signalr.NewServer(context.Background(),
		signalr.SimpleHubFactory(hub),
		signalr.HTTPTransports("ServerSentEvents"),
		signalr.KeepAliveInterval(2*time.Second),
		signalr.TimeoutInterval(10*time.Second),
		signalr.Logger(log.NewLogfmtLogger(os.Stdout), true),
		signalr.EnableDetailedErrors(true))

	cleanerTimer := time.NewTicker(5 * time.Minute)

	go func() {
		for range cleanerTimer.C {
			now := time.Now()
			for _, game := range games {
				if game.LastInteraction.Add(10 * time.Minute).Before(now) {
					//TODO: signalr remove from group
					delete(games, game.Id)
				}
			}
		}
	}()

	router := http.NewServeMux()

	corsMiddleware := &CorsMiddleware{Handler: router}

	server.MapHTTP(signalr.WithHTTPServeMux(router), "/hub")

	router.Handle("/api/games/", &gameHandler{
		signalrServer: hub,
	})
	router.Handle("/api/health", &healthHandler{})

	err = http.ListenAndServe(":8080", corsMiddleware)
	if err != nil {
		panic(err)
	}
}
