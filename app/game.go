package main

import (
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Handles the end of the game and sends the appropriate message.
func HandleGameEnd(s *discordgo.Session, game *MinesweeperGame, event int) {
	// Calculate the time taken in the game and format it as a human-readable string.
	timeString := fmt.Sprintf(
		"\nYour time was %s",
		humanizetime.HumanizeDuration(time.Now().Sub(game.StartTime), 3),
	)

	if game.StartTime.IsZero() {
		timeString = "\nyou never even started the game.."
	}

	// Determine the content string based on the event that caused the game to end.
	content := fmt.Sprintf("<@!%s> ", game.UserID)
	switch event {
	case minesweeper.ManualEnd:
		content += "you ended the game, wimp."
	case minesweeper.TimedEnd:
		content += "bro there is no way it takes you that long."
	case minesweeper.Won:
		content += "You won, woohoo."
	case minesweeper.Lost:
		content += "you fuckin lost LOL"
	}

	// Send a message to the channel with the game result and time information.
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

	// Update the game board message with the final state of the board.
	boardContent := "Game over. LOL."
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    &boardContent,
		Components: GenerateBoard(game, false, true),
		ID:         game.BoardID,
		Channel:    game.ChannelID,
	})
	if err != nil {
		fmt.Println(err)
	}

	// Remove the game from the active games map.
	delete(Games, game.UserID)
}

// Generates the message components for the game board.
func GenerateBoard(game *MinesweeperGame, firstGen, useSpotTypes bool) []discordgo.MessageComponent {
	var rows []discordgo.MessageComponent
	var currentRow discordgo.ActionsRow

	// Iterate through the spots on the game board.
	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			spot := game.Game.FindSpot(x, y)

			// Create a button component for the spot.
			button := &discordgo.Button{
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("boardx%dy%d", spot.X, spot.Y),
				Disabled: firstGen,
			}

			typeToUse := spot.DisplayedType
			if useSpotTypes {
				typeToUse = spot.Type
			}

			// Set the properties of the button based on the spot type.
			switch typeToUse {
			case minesweeper.Hidden:
				button.Emoji = discordgo.ComponentEmoji{
					Name: "invie",
					ID:   "1112567785076305971",
				}

			case minesweeper.Bomb:
				button.Emoji = discordgo.ComponentEmoji{
					Name: "💥",
				}
				button.Style = discordgo.DangerButton

			case minesweeper.Flag:
				button.Emoji = discordgo.ComponentEmoji{
					Name: "🚩",
				}
				button.Style = discordgo.SuccessButton

			case minesweeper.Normal:
				button.Style = discordgo.SecondaryButton
				button.Label = strconv.Itoa(spot.NearbyBombs)

			case minesweeper.StartHere:
				button.Emoji = discordgo.ComponentEmoji{
					Name: "clickme",
					ID:   "1119511692825604096",
				}
				button.Disabled = false
			}

			currentRow.Components = append(currentRow.Components, button)

			// Check if the current row is full and add it to the rows slice.
			if len(currentRow.Components) >= 5 {
				rows = append(rows, currentRow)
				currentRow = discordgo.ActionsRow{}
			}
		}
	}

	return rows
}
