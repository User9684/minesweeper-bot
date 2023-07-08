package minesweeper

import (
	"fmt"
	"math/rand"
)

// Spot types.
const (
	Hidden = iota
	Normal
	Bomb
	Flag
	StartHere
)

// Difficulties.
const (
	Easy = iota
	Medium
	Hard
)

// Outcomes.
const (
	Nothing = iota
	ManualEnd
	TimedEnd
	Lost
	Won
)

type Spot struct {
	X                int
	Y                int
	Type             int
	DisplayedType    int
	NearbyBombs      int
	SurroundingSpots []*Spot
}

type Game struct {
	Spots        map[string]*Spot
	VisitedZeros map[string]bool
	Difficulty   int
	SpotsLeft    int
	TotalBombs   int
}

func NewGame(dif int) *Game {
	spots, bombCount := generateSpots(dif)
	game := &Game{
		Spots:        spots,
		VisitedZeros: make(map[string]bool),
		Difficulty:   dif,
		SpotsLeft:    (5 * 5) - bombCount,
		TotalBombs:   bombCount,
	}

	return game
}

func (g *Game) VisitSpot(s *Spot) (bool, int) {
	if s.DisplayedType != Hidden && s.DisplayedType != StartHere {
		return false, Nothing
	}
	if s.Type == Bomb {
		s.DisplayedType = Bomb
		return true, Lost
	}

	// Update the displayed type of the spot.
	s.DisplayedType = Normal
	g.SpotsLeft--

	if g.SpotsLeft <= 0 {
		return true, Won
	}

	wonGame := false
	if s.NearbyBombs == 0 {
		wonGame = g.VisitNearbyZeros(s)
	}
	if wonGame {
		return true, Won
	}

	return false, Nothing
}

func (g *Game) FlagSpot(s *Spot) {
	// Set displayed type to hidden if already flagged.
	if s.DisplayedType == Flag {
		s.DisplayedType = Hidden
		return
	}
	// Prevent the flagging of already visited spots.
	if s.DisplayedType != Hidden {
		return
	}

	// Update the displayed type of the spot to Flag.
	s.DisplayedType = Flag
}

// Visits all the nearby zeros and then the zeros nearby those zeros, just like in real minesweeper.
func (g *Game) VisitNearbyZeros(s *Spot) bool {
	// Add the current spot to the list of visited zeros.
	g.VisitedZeros[getKey(s.X, s.Y)] = true

	// Visit the nearby spots recursively.
	for _, surroundingSpot := range s.SurroundingSpots {
		if surroundingSpot.Type != Normal {
			continue
		}
		if surroundingSpot.DisplayedType == Flag {
			continue
		}
		if g.HasVisitedZero(surroundingSpot) {
			continue
		}

		_, outcome := g.VisitSpot(surroundingSpot)
		if outcome == Won {
			return true
		}
	}

	return false
}

func (g *Game) HasVisitedZero(s *Spot) bool {
	return g.VisitedZeros[getKey(s.X, s.Y)]
}

// Visits the surrounding 8 spots if the surrounding flag count is equal to the number of surrounding bombs.
func (g *Game) ChordSpot(s *Spot) int {
	var spotsToVisit []*Spot
	flaggedCells := 0
	for _, spot := range s.SurroundingSpots {
		if spot.DisplayedType == Flag {
			flaggedCells++
			continue
		}
		spotsToVisit = append(spotsToVisit, spot)
	}
	if flaggedCells != s.NearbyBombs {
		return Nothing
	}
	for _, spot := range spotsToVisit {
		gameEnd, result := g.VisitSpot(spot)
		if gameEnd {
			return result
		}
	}

	return Nothing
}

func (g *Game) FindSpot(X, Y int) *Spot {
	key := getKey(X, Y)
	return g.Spots[key]
}

// Generates spots for the game to use.
func generateSpots(diff int) (map[string]*Spot, int) {
	Spots := make(map[string]*Spot)

	// Generate random start position.
	sx := rand.Intn(5)
	sy := rand.Intn(5)
	startPositionKey := getKey(sx, sy)

	ignoredPositions := map[string]bool{}
	bsx := sx - 1
	bsy := sy - 1
	for y := 0; y <= 2; y++ {
		for x := 0; x <= 2; x++ {
			newx := bsx + x
			newy := bsy + y
			if newx >= 5 || newx < 0 {
				continue
			}
			if newy >= 5 || newy < 0 {
				continue
			}
			newIgnoredKey := getKey(newx, newy)
			ignoredPositions[newIgnoredKey] = true
		}
	}

	ignoredPositions[startPositionKey] = true

	// Generate bomb positions.
	bombPositions := make(map[string]bool)
	for i := 0; i <= 4+(diff*2); i++ {
		// Generate position.
		key := getKey(rand.Intn(5), rand.Intn(5))

		// Repeat until position is valid, or has tried 10 times.
		for tries := 0; (bombPositions[key] || ignoredPositions[key]) && tries <= 9; tries++ {
			key = getKey(rand.Intn(5), rand.Intn(5))
		}

		if bombPositions[key] || ignoredPositions[key] {
			continue
		}
		bombPositions[key] = true
	}

	// Create spot instances.
	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			spot := Spot{
				X:                x,
				Y:                y,
				Type:             Normal,
				NearbyBombs:      0,
				DisplayedType:    Hidden,
				SurroundingSpots: []*Spot{},
			}
			key := getKey(x, y)

			if bombPositions[key] {
				spot.Type = Bomb
			}
			Spots[key] = &spot
		}
	}

	// Generate SurroundingSpots.
	for _, spot := range Spots {
		baseX := spot.X - 1
		baseY := spot.Y - 1

		for y := 0; y <= 2; y++ {
			for x := 0; x <= 2; x++ {
				newx := baseX + x
				newy := baseY + y
				if newx == spot.X && newy == spot.Y {
					continue
				}
				if newx >= 5 || newx < 0 {
					continue
				}
				if newy >= 5 || newy < 0 {
					continue
				}

				foundSpotKey := getKey(newx, newy)
				foundSpot := Spots[foundSpotKey]

				if foundSpot.Type == Bomb {
					spot.NearbyBombs++
				}

				spot.SurroundingSpots = append(spot.SurroundingSpots, foundSpot)
			}
		}
	}

	Spots[startPositionKey].DisplayedType = StartHere

	return Spots, len(bombPositions)
}

// Generates a map key for coordinate pair.
func getKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}
