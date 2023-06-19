package main

import (
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

var RequestOption = func(cfg *discordgo.RequestConfig) {

}

var InteractionHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	// Commands
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		content := "PONG"
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
		if err != nil {
			cmdError(i, err)
			return
		}

		m, err := s.InteractionResponse(i.Interaction, RequestOption)
		if err != nil {
			cmdError(i, err)
			return
		}

		startID, err := strconv.Atoi(i.ID)
		if err != nil {
			cmdError(i, err)
			return
		}
		endID, err := strconv.Atoi(m.ID)
		if err != nil {
			cmdError(i, err)
			return
		}

		var (
			startTS = int64(startID) >> int64(22)
			endTS   = int64(endID) >> int64(22)
		)

		content = fmt.Sprintf(
			"Heartbeat Latency: `%d`ms\nHTTP Latency:`%d`ms",
			s.HeartbeatLatency().Milliseconds(),
			endTS-startTS,
		)

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		}); err != nil {
			cmdError(i, err)
			return
		}
	},
	"minesweeper": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Convert options to map
		optionMap := mapOptions(i.ApplicationCommandData().Options)
		userID, isGuild := getUserID(i)

		if game, ok := Games[userID]; ok {
			var (
				location  = i.GuildID
				channel   = i.ChannelID
				messageID = game.BoardID
			)
			if !isGuild {
				location = "@me"
				channel = s.State.User.ID
			}

			replyContent := fmt.Sprintf("You already have a game open at https://discord.com/channels/%s/%s/%s", location, channel, messageID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			})
			return
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		var Game *minesweeper.Game

		switch optionMap["difficulty"].Value {
		case "easy":
			Game = minesweeper.NewGame(minesweeper.Easy)
		case "medium":
			Game = minesweeper.NewGame(minesweeper.Medium)
		case "hard":
			Game = minesweeper.NewGame(minesweeper.Hard)
		}

		newGame := MinesweeperGame{
			UserID:     userID,
			GuildID:    i.GuildID,
			ChannelID:  i.ChannelID,
			Game:       Game,
			Difficulty: fmt.Sprintf("%v", optionMap["difficulty"].Value),
		}

		content := "Click the <:clickme:1119511692825604096> to start the game!"
		board := GenerateBoard(&newGame, true, false)

		msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:    &content,
			Components: &board,
		})
		if err != nil {
			cmdError(i, err)
			return
		}

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

		flagMsg, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
			Reference: msg.Reference(),
			Components: []discordgo.MessageComponent{
				flagRow,
			},
		}, RequestOption)
		if err != nil {
			cmdError(i, err)
			return
		}

		newGame.BoardID = msg.ID
		newGame.FlagID = flagMsg.ID

		Games[userID] = &newGame
	},
	"admin": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		userID, _ := getUserID(i)

		if a, ok := Admins[userID]; !a || !ok {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: "No.",
				},
			}); err != nil {
				cmdError(i, err)
			}

			return
		}
		subcommand := i.Interaction.ApplicationCommandData().Options[0]
		optionMap := mapOptions(subcommand.Options)

		switch subcommand.Name {
		case "blacklist":
			target := optionMap["user"].UserValue(s).ID
			var message string
			if msg, ok := optionMap["message"]; ok {
				message = msg.StringValue()
			}

			if message == "" {
				message = "No message provided"
			}

			blacklistUser(target, message)

			replyContent := fmt.Sprintf("Blacklisted `%s` for reason: `%s`", target, message)
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			}); err != nil {
				cmdError(i, err)
			}

		case "unblacklist":
			target := optionMap["user"].UserValue(s).ID

			unblacklistUser(target)

			replyContent := fmt.Sprintf("Removed blacklist for `%s`", target)
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			}); err != nil {
				cmdError(i, err)
			}

		case "leaderboardmsg":
			// Not setup yet, will exist in the next commit.
		}
	},

	// Components.
	"minesweeperflagbutton": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		userID, _ := getUserID(i)
		game, ok := Games[userID]
		if !ok {
			replyContent := "You don't have a game open!"
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			})
			return
		}

		if game.FlagID != i.Message.ID {
			replyContent := "This is not your game."
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			})
			return
		}

		// Response to interaction.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		// Toggle flag.
		game.flagEnabled = !game.flagEnabled

		// Create the new button.
		flagRow := discordgo.ActionsRow{}
		flagButton := &discordgo.Button{
			CustomID: "minesweeperflagbutton",
			Style:    discordgo.DangerButton,
			Label:    "OFF",
			Emoji: discordgo.ComponentEmoji{
				Name: "🚩",
			},
		}
		endGameButton := &discordgo.Button{
			CustomID: "endgamebutton",
			Style:    discordgo.DangerButton,
			Label:    "End game",
		}

		if game.flagEnabled {
			flagButton.Label = "ON"
			flagButton.Style = discordgo.SuccessButton
		}
		flagRow.Components = append(flagRow.Components, flagButton, endGameButton)

		// Edit old flag message with new button.
		if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    game.ChannelID,
			ID:         game.FlagID,
			Components: []discordgo.MessageComponent{flagRow},
		}); err != nil {
			cmdError(i, err)
			return
		}
	},
	"endgamebutton": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Response to interaction.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		userID, _ := getUserID(i)

		game, ok := Games[userID]
		if !ok {
			replyContent := "You don't have a game open!"
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: replyContent,
				},
			})
			return
		}

		HandleGameEnd(s, game, minesweeper.Nothing)
	},
}

func HandleBoard(s *discordgo.Session, i *discordgo.InteractionCreate, positionx, positiony int) {
	userID, _ := getUserID(i)
	game, ok := Games[userID]
	if !ok {
		replyContent := "You don't have a game open!"
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: replyContent,
			},
		})
		return
	}

	if game.BoardID != i.Message.ID {
		replyContent := "This is not your game."
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: replyContent,
			},
		})
		return
	}

	if game.StartTime.IsZero() {
		game.StartTime = time.Now()
	}

	spot := game.Game.FindSpot(positionx, positiony)

	switch game.flagEnabled {
	case true:
		game.Game.FlagSpot(spot)

	case false:
		gameEnd, event := game.Game.VisitSpot(spot)
		if gameEnd {
			HandleGameEnd(s, game, event)
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			return
		}
	}

	content := "here you go >~<"
	board := GenerateBoard(game, false, false)

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    game.ChannelID,
		ID:         game.BoardID,
		Content:    &content,
		Components: board,
	}); err != nil {
		cmdError(i, err)
		return
	}

	go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

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

func CmdInit(s *discordgo.Session) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))
	for i, v := range Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			fmt.Printf("Cannot create '%v' command: %v\n", v.Name, err)
		}
		registeredCommands[i] = cmd
	}
}

func mapOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	var optionMap = make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}

func getUserID(i *discordgo.InteractionCreate) (string, bool) {
	var inGuild bool
	var userID string
	if i.User == nil {
		userID = i.Member.User.ID
		inGuild = true
	}
	if i.Member == nil {
		userID = i.User.ID
		inGuild = false
	}

	return userID, inGuild
}

func cmdError(i *discordgo.InteractionCreate, err error) {
	if err == nil {
		return
	}
	fmt.Println(err)

	go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("An error occured! \n```%s```", err.Error()),
		},
	})
}
