package main

import (
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleGameEnd(s *discordgo.Session, game *MinesweeperGame, event int) {
	timeString := fmt.Sprintf(
		"\nYour time was %s",
		humanizetime.HumanizeDuration(time.Now().Sub(game.StartTime), 3),
	)

	if game.StartTime.IsZero() {
		timeString = "\nYou did not start the game."
	}

	var content string
	// Game end string based off of end cause.
	switch event {
	case minesweeper.Nothing:
		content = "game ended lol"
	case minesweeper.Won:
		content = "you won woohoo"
	case minesweeper.Lost:
		content = "you fuckin lost LOL"
	}

	_, err := s.ChannelMessageSendComplex(game.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("%s%s", content, timeString),
		Reference: &discordgo.MessageReference{
			MessageID: game.BoardID,
			ChannelID: game.ChannelID,
			GuildID:   game.GuildID,
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	c := "game over lol"
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &c,
		Components: GenerateBoard(game, false, true),
		ID:         game.BoardID,
		Channel:    game.ChannelID,
	})
	if err != nil {
		fmt.Println(err)
	}

	delete(Games, game.UserID)
}

func GenerateBoard(Game *MinesweeperGame, FirstGen, useSpotTypes bool) []discordgo.MessageComponent {
	var Rows []discordgo.MessageComponent
	currentRow := discordgo.ActionsRow{}

	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			spot := Game.Game.FindSpot(x, y)

			// Create the button.
			newComponent := &discordgo.Button{
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("boardx%dy%d", spot.X, spot.Y),
			}

			if FirstGen {
				newComponent.Disabled = true
			}

			typeToUse := spot.DisplayedType
			if useSpotTypes {
				typeToUse = spot.Type
			}

			switch typeToUse {
			case minesweeper.Hidden:
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "invie",
					ID:   "1112567785076305971",
				}

			case minesweeper.Bomb:
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "💥",
				}
				newComponent.Style = discordgo.DangerButton

			case minesweeper.Flag:
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "🚩",
				}
				newComponent.Style = discordgo.SuccessButton

			case minesweeper.Normal:
				newComponent.Style = discordgo.SecondaryButton
				newComponent.Label = strconv.Itoa(spot.NearbyBombs)
			case minesweeper.StartHere:
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "clickme",
					ID:   "1119511692825604096",
				}
				newComponent.Disabled = false
			}

			currentRow.Components = append(currentRow.Components, newComponent)

			if len(currentRow.Components) >= 5 {
				Rows = append(Rows, currentRow)
				currentRow = discordgo.ActionsRow{}
			}
		}
	}

	return Rows
}
