package main

import (
	"fmt"
	"main/minesweeper"
	"math"

	"github.com/bwmarrin/discordgo"
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
		Description: "Tried to chord, but you were out of tune.",
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
	5: {
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

func getFieldsAndComponents(userData UserData, page int) ([]*discordgo.MessageEmbedField, []discordgo.MessageComponent) {
	var components []discordgo.MessageComponent
	var fields []*discordgo.MessageEmbedField

	pages := math.Ceil(float64(len(userData.Achievements)) / float64(5))

	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: "profileleft",
				Label:    "←",
				Style:    discordgo.PrimaryButton,
				Disabled: page <= 0,
			},
			discordgo.Button{
				CustomID: "profileright",
				Label:    "→",
				Style:    discordgo.PrimaryButton,
				Disabled: float64(page+1) >= pages,
			},
		},
	})

	paginated := paginate(userData.Achievements, page, 5)
	fmt.Println(paginated)
	for _, ID := range paginated {
		achievement := Achievements[ID]
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  achievement.Name,
			Value: achievement.Description,
		})
	}

	if len(userData.Achievements) <= 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: "No unlocked achievements",
		})
	}

	return fields, components
}

func paginate(numbers []int, page, itemsPerPage int) []int {
	pages := math.Ceil(float64(len(numbers)) / float64(itemsPerPage))
	if float64(page) >= pages {
		return []int{}
	}

	startIndex := page * itemsPerPage
	endIndex := int(math.Min(float64(startIndex+itemsPerPage), float64(len(numbers))))

	paginatedList := numbers[startIndex:endIndex]

	return paginatedList
}
