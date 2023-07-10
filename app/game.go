package main

import (
	"context"
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StartGame(s *discordgo.Session, i *discordgo.InteractionCreate, Game *minesweeper.Game, Difficulty, userID string) {
	// Create a new MinesweeperGame object to store game information.
	newGame := MinesweeperGame{
		UserID:     userID,
		GuildID:    i.GuildID,
		ChannelID:  i.ChannelID,
		Game:       Game,
		Difficulty: Difficulty,
	}

	// Send the initial message with the game board.
	content := "Click the <:clickme:1119511692825604096> to start the game!"
	board := GenerateBoard(&newGame, true, false)
	msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &board,
	})
	if err != nil {
		cmdError(s, i, err)
		return
	}

	// Add flag and end game buttons to the message.
	flagRow := discordgo.ActionsRow{}
	flagRow.Components = append(flagRow.Components, &discordgo.Button{
		CustomID: "minesweeperflagbutton",
		Style:    discordgo.DangerButton,
		Label:    "OFF",
		Emoji: discordgo.ComponentEmoji{
			Name: "🚩",
		},
	}, &discordgo.Button{
		CustomID: "endgamebutton",
		Style:    discordgo.DangerButton,
		Label:    "End game",
	})

	// Send the flag and end game buttons as a separate message.
	flagMsg, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
		Reference:  msg.Reference(),
		Components: []discordgo.MessageComponent{flagRow},
	}, RequestOption)
	if err != nil {
		cmdError(s, i, err)
		return
	}

	// Update the new game object with message IDs.
	newGame.BoardID = msg.ID
	newGame.FlagID = flagMsg.ID

	// Configure automatic end game timer.
	timer := time.NewTimer(time.Duration(EndAfter) * time.Second)
	channel := make(chan struct{})
	newGame.EndGameChan = &channel

	// Start automatic end game timer.
	if EndAfter != 0 {
		go func() {
			for {
				select {
				case <-timer.C:
					HandleGameEnd(s, &newGame, minesweeper.TimedEnd, false)
					return
				case <-channel:
					timer.Stop()
					return
				}
			}
		}()
	}

	// Store the new game object in the Games map.
	Games[userID] = &newGame
}

// Handles the end of the game and sends the appropriate message.
func HandleGameEnd(s *discordgo.Session, game *MinesweeperGame, event int, addToBoard bool) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	close(*game.EndGameChan)
	// Calculate the time taken in the game and format it as a human-readable string.
	gameDuration := time.Since(game.StartTime)
	timeString := fmt.Sprintf(
		"\nYour time was %s",
		humanizetime.HumanizeDuration(gameDuration, 3),
	)

	if game.StartTime.IsZero() {
		timeString = "\nyou never even started the game.."
	}

	boardContent := ""

	// Determine the content string based on the event that caused the game to end.
	content := fmt.Sprintf("<@!%s> ", game.UserID)
	switch event {
	case minesweeper.ManualEnd:
		content += getRandomMessage(SarcasticGiveUpMessages)
		boardContent = "LOL, giving up already?"
	case minesweeper.TimedEnd:
		content += getRandomMessage(SarcasticTimeOverMessages)
		boardContent = "Jesus christ, it does NOT take that long!"
	case minesweeper.Lost:
		boardContent = "Game over. LOL."
		content += getRandomMessage(SarcasticLostMessages)

		if game.Difficulty == "custom" {
			break
		}

		userData := getUserData(game.UserID)

		dd := userData.Difficulties[game.Difficulty]

		dd.Losses++

		userData.Difficulties[game.Difficulty] = dd

		filter := bson.D{{
			Key:   "userID",
			Value: game.UserID,
		}}

		data, err := bson.Marshal(userData)
		if err != nil {
			fmt.Println(err)
			return
		}

		var update bson.M
		if err := bson.Unmarshal(data, &update); err != nil {
			return
		}

		request := d.Collection("userdata").FindOneAndUpdate(
			context.TODO(),
			filter,
			bson.D{{
				Key:   "$set",
				Value: update,
			}},
			options.FindOneAndUpdate().SetUpsert(true),
		)

		if err := request.Decode(&userData); err != nil {
			fmt.Println(err)
		}
	case minesweeper.Won:
		boardContent = "Wow, you managed to win!"
		game.Won = true

		messages := getMessages(int64(gameDuration.Seconds()))
		if gameDuration.Seconds() <= 0.2 {
			messages = SarcasticOneClickMessages
			addToBoard = false
		}

		content += getRandomMessage(messages)
		if !addToBoard {
			break
		}
		if game.Difficulty == "custom" {
			break
		}
		entry := LeaderboardEntry{
			Time:   gameDuration.Seconds(),
			UserID: game.UserID,
			Spot:   11,
		}
		if game.GuildID != "" {
			addToLeaderboard(game.GuildID, game.Game.Difficulty, entry)
		}
		addToLeaderboard("global", game.Game.Difficulty, entry)

		userData := getUserData(game.UserID)

		dd := userData.Difficulties[game.Difficulty]

		dd.Wins++
		if dd.PB > gameDuration.Seconds() || dd.PB == 0 {
			dd.PB = gameDuration.Seconds()
		}
		if dd.PW < gameDuration.Seconds() {
			dd.PW = gameDuration.Seconds()
		}

		if userData.Difficulties == nil {
			userData.Difficulties = map[string]DifficultyData{}
		}
		userData.Difficulties[game.Difficulty] = dd

		filter := bson.D{{
			Key:   "userID",
			Value: game.UserID,
		}}

		data, err := bson.Marshal(userData)
		if err != nil {
			fmt.Println(err)
			return
		}

		var update bson.M
		if err := bson.Unmarshal(data, &update); err != nil {
			return
		}

		request := d.Collection("userdata").FindOneAndUpdate(
			context.TODO(),
			filter,
			bson.D{{
				Key:   "$set",
				Value: update,
			}},
			options.FindOneAndUpdate().SetUpsert(true),
		)

		if err := request.Decode(&userData); err != nil {
			fmt.Println(err)
		}
	}
	boardContent += fmt.Sprintf("\n<@!%s>'s **%s** minesweeper game", game.UserID, strings.ToUpper(game.Difficulty))

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
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
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

			if useSpotTypes && spot.DisplayedType == minesweeper.Flag && spot.Type == minesweeper.Bomb {
				typeToUse = minesweeper.Flag
			}

			if game.Won && spot.Type == minesweeper.Bomb {
				typeToUse = minesweeper.Flag
			}

			// Set the properties of the button based on the spot type.
			switch typeToUse {
			case minesweeper.Hidden:
				button.Emoji = discordgo.ComponentEmoji{
					Name: "invie",
					ID:   "1112567785076305971",
				}

			case minesweeper.Normal:
				button.Style = discordgo.SecondaryButton
				button.Label = strconv.Itoa(spot.NearbyBombs)

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
