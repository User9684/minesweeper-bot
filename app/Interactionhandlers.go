package main

import (
	"fmt"
	"main/minesweeper"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var RequestOption = func(cfg *discordgo.RequestConfig) {

}

var InteractionHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"minesweeper": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Convert options to map
		optionMap := mapOptions(i.ApplicationCommandData().Options)
		userID, isGuild := getUserID(i)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

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
			GuildID:    i.GuildID,
			ChannelID:  i.ChannelID,
			Game:       Game,
			Difficulty: fmt.Sprintf("%v", optionMap["difficulty"].Value),
		}

		content := "here you go >~<"
		board := GenerateBoard(&newGame)

		msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content:    &content,
			Components: &board,
		})
		if err != nil {
			fmt.Println(err)
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
		})

		flagMsg, err := s.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{
			Reference: msg.Reference(),
			Components: []discordgo.MessageComponent{
				flagRow,
			},
		}, RequestOption)
		if err != nil {
			fmt.Println(err)
			return
		}

		newGame.BoardID = msg.ID
		newGame.FlagID = flagMsg.ID

		Games[userID] = &newGame
	},
	"minesweeperflagbutton": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Response to interaction.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		userID, _ := getUserID(i)
		game := Games[userID]

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
		if game.flagEnabled {
			flagButton.Label = "ON"
			flagButton.Style = discordgo.SuccessButton
		}
		flagRow.Components = append(flagRow.Components, flagButton)

		// Edit old flag message with new button.
		_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    game.ChannelID,
			ID:         game.FlagID,
			Components: []discordgo.MessageComponent{flagRow},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func HandleBoard(s *discordgo.Session, i *discordgo.InteractionCreate, positionx, positiony int) {
	userID, _ := getUserID(i)
	game := Games[userID]

	spot := game.Game.FindSpot(positionx, positiony)

	if game.flagEnabled {
		game.Game.FlagSpot(spot)
	}

	if !game.flagEnabled {
		game.Game.VisitSpot(spot)
	}

	content := "here you go >~<"
	board := GenerateBoard(game)

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    game.ChannelID,
		ID:         game.BoardID,
		Content:    &content,
		Components: board,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

func GenerateBoard(Game *MinesweeperGame) []discordgo.MessageComponent {
	var Rows []discordgo.MessageComponent
	currentRow := discordgo.ActionsRow{}

	for y := 0; y <= 4; y++ {
		for x := 0; x <= 4; x++ {
			spot := Game.Game.FindSpot(x, y)

			// Create the button.
			newComponent := &discordgo.Button{
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("boardx%dy%d", spot.X, spot.Y),
			}

			switch spot.DisplayedType {
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

			default:
				newComponent.Label = strconv.Itoa(spot.NearbyBombs)
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

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("An error occured! \n```%s```", err.Error()),
		},
	})
}
