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
)

// Difficulties.
const (
	Easy = iota
	Medium
	Hard
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
}

func NewGame(dif int) *Game {
	game := &Game{
		Spots:        generateSpots(dif),
		Difficulty:   dif,
		VisitedZeros: make(map[string]bool),
	}

	return game
}

func (g *Game) VisitSpot(s *Spot) bool {
	if s.Type == Flag {
		return false
	}
	if s.Type == Bomb {
		s.DisplayedType = Bomb
		return true
	}

	// Update the displayed type of the spot.
	s.DisplayedType = Normal

	if s.NearbyBombs == 0 {
		g.VisitNearbyZeros(s)
	}

	return false
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
		if surroundingSpot.Type == Bomb {
			continue
		}
		if surroundingSpot.Type != Hidden {
			continue
		}

		if g.HasVisitedZero(surroundingSpot) {
			continue
		}

		g.VisitSpot(surroundingSpot)
		if surroundingSpot.NearbyBombs == 0 {
			g.VisitNearbyZeros(surroundingSpot)
		}
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
func generateSpots(diff int) map[string]*Spot {
	Spots := make(map[string]*Spot)

	// Generate bomb positions.
	bombPositions := make(map[string]bool)
	for i := 0; i <= 4+(diff*2); i++ {
		// Generate position.
		x, _ := rand.Int(rand.Reader, big.NewInt(4))
		y, _ := rand.Int(rand.Reader, big.NewInt(4))

		key := getKey(int(x.Int64()), int(y.Int64()))
		bombPositions[key] = true
	}

	// Create spot instances.
	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			spot := Spot{
				X:                x,
				Y:                y,
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

	return Spots
}

// Generates a map key for coordinate pair.
func getKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}
