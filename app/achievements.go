package main

import (
	"main/minesweeper"
)

type CheckData struct {
	Event       int
	ClickedCell *minesweeper.Spot
	Chorded     bool
	PreVisit    bool
}

type Achievement struct {
	Name        string
	Description string
	CheckFunc   func(data CheckData) bool
}

var Achievements = map[int]Achievement{
	// Win a game.
	0: {
		Name:        "Basic reasoning",
		Description: "Win your first game of minesweeper",
		CheckFunc: func(data CheckData) bool {
			return data.Event == minesweeper.Won
		},
	},
	// Lose a game.
	1: {
		Name:        "U MAD?",
		Description: "Lose your first game of minesweeper",
		CheckFunc: func(data CheckData) bool {
			return data.Event == minesweeper.Lost
		},
	},
	// Failed to chord.
	2: {
		Name:        "Out Of Tune",
		Description: "You tried to chord, but you were out of tune.",
		CheckFunc: func(data CheckData) bool {
			if !data.Chorded {
				return false
			}
			return data.Event == minesweeper.Lost
		},
	},
	// Successfully chord.
	3: {
		Name:        "Wonderful music!",
		Description: "Successfully chord",
		CheckFunc: func(data CheckData) bool {
			if !data.Chorded {
				return false
			}
			return data.Event != minesweeper.Lost
		},
	},
	// Randomly click the board without any neighboring cells.
	4: {
		Name:        "YOLO",
		Description: "Clicked a cell without any surrounding visited cells",
		CheckFunc: func(data CheckData) bool {
			if data.ClickedCell.DisplayedType == minesweeper.StartHere {
				return false
			}
			for _, spot := range data.ClickedCell.SurroundingSpots {
				if spot.DisplayedType != minesweeper.Hidden {
					return false
				}
			}
			return true
		},
	},
	69: {
		Name:        "⬧︎♓︎●︎●︎⍓︎ ♍︎♋︎⧫︎ ⬧︎♋︎⍓︎⬧︎ ♒︎♓︎✏︎",
		Description: "🖳︎🗏︎",
		CheckFunc: func(data CheckData) bool {
			return false
		},
	},
}

func AwardAchievements(game *MinesweeperGame, event int, clickedCell *minesweeper.Spot, chord, beforeVisit bool) map[int]Achievement {
	var achievementsGotten = make(map[int]Achievement)

	data := CheckData{
		Event:       event,
		ClickedCell: clickedCell,
		Chorded:     chord,
		PreVisit:    beforeVisit,
	}

	for ID, achievment := range Achievements {
		if achievment.CheckFunc(data) {
			achievementsGotten[ID] = achievment
		}
	}

	return achievementsGotten
}
