package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/rrborja/minesweeper"
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

		game, _ := minesweeper.NewGame(minesweeper.Grid{
			Width:  5,
			Height: 5,
		})

		switch optionMap["difficulty"].Value {
		case "easy":
			game.SetDifficulty(1)
		case "medium":
			game.SetDifficulty(2)
		case "hard":
			game.SetDifficulty(3)
		}

		newGame := MinesweeperGame{
			GuildID:    i.GuildID,
			ChannelID:  i.ChannelID,
			Game:       game,
			Difficulty: fmt.Sprintf("%v", optionMap["difficulty"].Value),
		}

		game.Play()

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

		// Response to interaction.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	},
}

func HandleBoard(s *discordgo.Session, i *discordgo.InteractionCreate, positionx, positiony int) {
	userID, _ := getUserID(i)
	game := Games[userID]

	if game.flagEnabled {
		flagSpot(game, positionx, positiony)
	}

	if !game.flagEnabled {
		visited, err := game.Game.Visit(positionx, positiony)
		if err != nil {
			fmt.Println(err)
		}
		game.VisitedCells = append(game.VisitedCells, visited...)
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

func flagSpot(game *MinesweeperGame, x, y int) {
	game.Game.Flag(x, y)
	for i, spot := range game.FlaggedCells {
		if spot.X != x {
			continue
		}
		if spot.Y != y {
			continue
		}

		game.FlaggedCells = append(game.FlaggedCells[:i], game.FlaggedCells[i+1:]...)
		return
	}

	game.FlaggedCells = append(game.FlaggedCells, struct {
		X int
		Y int
	}{
		X: x,
		Y: y,
	})
}

func GenerateBoard(Game *MinesweeperGame) []discordgo.MessageComponent {
	var Rows []discordgo.MessageComponent

	for y := 0; y <= 4; y++ {
		currentRow := discordgo.ActionsRow{}
		for x := 0; x <= 4; x++ {
			newComponent := &discordgo.Button{
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("boardx%dy%d", x, y),
			}
			str := GetBoardSpot(Game, x, y)

			switch str {
			case "inv":
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name:     "invie",
					ID:       "1112567785076305971",
					Animated: false,
				}
			case "exp":
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "💥",
				}
			case "flg":
				newComponent.Emoji = discordgo.ComponentEmoji{
					Name: "🚩",
				}
			default:
				newComponent.Label = str
			}

			currentRow.Components = append(currentRow.Components, newComponent)
		}
		Rows = append(Rows, currentRow)
	}

	return Rows
}

func GetBoardSpot(Game *MinesweeperGame, x, y int) string {
	for _, spot := range Game.VisitedCells {
		if spot.X() == x && spot.Y() == y {
			if spot.Flagged() {
				return "flg"
			}
			if spot.Visited() && (spot.Node == 2) {
				return "exp"
			}
			if spot.Visited() {
				return strconv.Itoa(spot.Value)
			}
		}
	}

	for _, spot := range Game.FlaggedCells {
		if spot.X == x && spot.Y == y {
			return "flg"
		}
	}

	return "inv"
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
