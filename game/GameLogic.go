package game

func (g *Game) IsValidMove(move Move) bool {
	return true
}

func (g *Game) IsValidPromotion(move Move) bool {
	return true
}

func (g *Game) IsCheckmate() bool {
	return false
}

func (g *Game) IsStalemate() bool {
	return false
}

func (g *Game) IsCheck() bool {
	return false
}
