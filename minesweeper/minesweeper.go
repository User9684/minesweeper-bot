package minesweeper

import (
	"crypto/rand"
	"fmt"
	"math/big"
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
}

func NewGame(dif int) *Game {
	spots, bombCount := generateSpots(dif)
	game := &Game{
		Spots:        spots,
		VisitedZeros: make(map[string]bool),
		Difficulty:   dif,
		SpotsLeft:    (5 * 5) - bombCount,
	}

	return game
}

func (g *Game) VisitSpot(s *Spot) (bool, int) {
	if s.Type == Flag {
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

	g.VisitNearbyZeros(s)

	return false, Nothing
}

func (g *Game) FlagSpot(s *Spot) {
	if s.DisplayedType != Hidden {
		return
	}

	// Update the displayed type of the spot to Flag.
	s.DisplayedType = Flag
}

// Visits all the nearby zeros and then the zeros nearby those zeros, just like in real minesweeper.
func (g *Game) VisitNearbyZeros(s *Spot) {
	if s.NearbyBombs != 0 {
		return
	}

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

		g.VisitSpot(surroundingSpot)
	}
}

func (g *Game) HasVisitedZero(s *Spot) bool {
	if g.VisitedZeros[getKey(s.X, s.Y)] {
		return true
	}
	return false
}

func (g *Game) FindSpot(X, Y int) *Spot {
	key := getKey(X, Y)
	return g.Spots[key]
}

// Generates spots for the game to use.
func generateSpots(diff int) (map[string]*Spot, int) {
	Spots := make(map[string]*Spot)

	// Generate random start position.
	sx, _ := rand.Int(rand.Reader, big.NewInt(4))
	sy, _ := rand.Int(rand.Reader, big.NewInt(4))
	startPositionKey := getKey(int(sx.Int64()), int(sy.Int64()))

	// Generate bomb positions.
	bombPositions := make(map[string]bool)
	for i := 0; i <= 4+(diff*2); i++ {
		// Generate position.
		x, _ := rand.Int(rand.Reader, big.NewInt(4))
		y, _ := rand.Int(rand.Reader, big.NewInt(4))
		key := getKey(int(x.Int64()), int(y.Int64()))

		// Repeat until position is valid, or has tried 5 times.
		tries := 0
		for (bombPositions[key] || key == startPositionKey) && tries <= 4 {
			// If this appears in console, don't worry. You only need to worry if it somehow is over 5.
			fmt.Printf("Invalid bomb position, retrying... (Retry attempt #%d)\n", tries)
			tries++
			x, _ = rand.Int(rand.Reader, big.NewInt(4))
			y, _ = rand.Int(rand.Reader, big.NewInt(4))
			key = getKey(int(x.Int64()), int(y.Int64()))
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

		index := 0
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

				//spot.SurroundingSpots[index] = foundSpot
				spot.SurroundingSpots = append(spot.SurroundingSpots, foundSpot)
				index++
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
