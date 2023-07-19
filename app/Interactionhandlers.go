package main

import (
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DifficultyLevel struct {
	PB string
	PW string
}

var RequestOption = func(cfg *discordgo.RequestConfig) {}
var difficultyOrder = []string{"easy", "medium", "hard"}

// Map command names to their respected handler.
var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Respond with "PONG" initially.
		content := "PONG"
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
			},
		})
		if err != nil {
			cmdError(s, i, err)
			return
		}

		// Get the response message.
		m, err := s.InteractionResponse(i.Interaction, RequestOption)
		if err != nil {
			cmdError(s, i, err)
			return
		}

		// Calculate latency and update the response message.
		startID, err := strconv.Atoi(i.ID)
		if err != nil {
			cmdError(s, i, err)
			return
		}
		endID, err := strconv.Atoi(m.ID)
		if err != nil {
			cmdError(s, i, err)
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

		// Update the response message with the calculated latency.
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		}); err != nil {
			cmdError(s, i, err)
			return
		}
	},
	"minesweeper": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Convert options to map.
		optionMap := mapOptions(i.ApplicationCommandData().Options)
		userID, isGuild := getUserID(i)

		// Check if the user already has a game open.
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

		// Respond with a deferred message update initially.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// Lock thread if a minesweeper command is being processed in this channel.
		// This is to prevent the situation shown in ../ai4aeISn.png.
		mutex, ok := ChannelMutex[i.ChannelID]
		if !ok {
			newMutex := &sync.Mutex{}
			mutex = newMutex
			ChannelMutex[i.ChannelID] = newMutex
		}
		mutex.Lock()
		defer mutex.Unlock()

		// Create a new game based on the selected difficulty.
		var Game *minesweeper.Game
		switch optionMap["difficulty"].Value {
		case "easy":
			Game = minesweeper.NewGame(minesweeper.Easy, 0, false, false)
		case "medium":
			Game = minesweeper.NewGame(minesweeper.Medium, 0, false, false)
		case "hard":
			Game = minesweeper.NewGame(minesweeper.Hard, 0, false, false)
		}

		StartGame(s, i, Game, fmt.Sprintf("%v", optionMap["difficulty"].Value), userID)
	},
	"custom": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()

		optionMap := mapOptions(i.ApplicationCommandData().Options)
		userID, _ := getUserID(i)

		bombs := optionMap["bombs"].IntValue()
		allowSurroundingBombs := optionMap["surroundingbombs"].BoolValue()
		noStartSpot := optionMap["nostartspot"].BoolValue()

		// Respond with a deferred message update initially.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		if bombs <= 0 {
			bombs = 1
		}
		if bombs >= 24 {
			bombs = 24
		}

		Game := minesweeper.NewGame(minesweeper.Custom, int(bombs), allowSurroundingBombs, noStartSpot)
		StartGame(s, i, Game, "custom", userID)
	},
	"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Respond with a deferred message update initially.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		optionMap := mapOptions(i.ApplicationCommandData().Options)
		_, isGuild := getUserID(i)

		targetGuild := "global"
		guildName := targetGuild

		if isGuild {
			targetGuild = i.GuildID
			guild, err := s.Guild(i.GuildID, RequestOption)
			if err != nil {
				cmdError(s, i, err)
				return
			}
			guildName = fmt.Sprintf("%s's", guild.Name)
		}

		if v, ok := optionMap["global"]; ok && v.BoolValue() {
			targetGuild = "global"
			guildName = targetGuild
		}

		embed, err := generateLeaderboardEmbed(targetGuild, guildName, optionMap["difficulty"].StringValue())
		if err != nil {
			cmdError(s, i, err)
			return
		}

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{&embed},
		}); err != nil {
			cmdError(s, i, err)
		}
	},
	"profile": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Respond with a deferred message update initially.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})

		// Convert options to map.
		optionMap := mapOptions(i.ApplicationCommandData().Options)
		userID, _ := getUserID(i)

		var targetID string
		target, ok := optionMap["target"]
		if !ok {
			targetID = userID
		}
		if ok {
			targetID = target.UserValue(s).ID
		}
		var viewAchievements bool
		view, ok := optionMap["achievements"]
		if !ok {
			viewAchievements = false
		}
		if ok {
			viewAchievements = view.BoolValue()
		}

		userData := getUserData(targetID)

		var userString string
		var userImage string
		user, err := getUser(targetID, false)
		if err != nil {
			fmt.Println(err)
			userString = targetID
			userImage = "https://cdn.discordapp.com/embed/avatars/0.png"
		}
		if err == nil {
			userString = user.Username
			userImage = user.AvatarURL("")
		}

		var fields []*discordgo.MessageEmbedField

		difficulties := map[string]*DifficultyLevel{
			"easy":   {PB: "Never played", PW: "Never played"},
			"medium": {PB: "Never played", PW: "Never played"},
			"hard":   {PB: "Never played", PW: "Never played"},
		}

		for level, data := range userData.Difficulties {
			pbd, err := time.ParseDuration(fmt.Sprintf("%fs", data.PB))
			if err == nil && pbd.Seconds() != 0 {
				difficulties[level].PB = humanizetime.HumanizeDuration(pbd, 3)
			}

			pwd, err := time.ParseDuration(fmt.Sprintf("%fs", data.PW))
			if err == nil && pwd.Seconds() != 0 {
				difficulties[level].PW = humanizetime.HumanizeDuration(pwd, 3)
			}
		}
		var title string
		var desc string
		if !viewAchievements {
			title = fmt.Sprintf("%s's Stats", userString)
			for _, difficulty := range difficultyOrder {
				difficultyData := difficulties[difficulty]
				fieldValue := fmt.Sprintf(UserStatsFormatString,
					userData.Difficulties[difficulty].Wins,
					userData.Difficulties[difficulty].Losses,
					userData.Difficulties[difficulty].WinStreak,
					difficultyData.PB,
					difficultyData.PW)

				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Stats for **%s** mode", strings.ToUpper(difficulty)),
					Value:  fieldValue,
					Inline: true,
				})
			}
		}
		var components []discordgo.MessageComponent
		if viewAchievements {
			title = fmt.Sprintf("%s's Achievements", userString)
			fields, components = getFieldsAndComponents(userData, 0)
			desc = "Page #1"
		}

		embed := discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       title,
			Description: desc,
			Color:       randomEmbedColor(),
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s - %s", userID, targetID),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: userImage,
			},
		}

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				&embed,
			},
			Components: &components,
		}); err != nil {
			fmt.Println(err)
		}
	},
	"admin": AdminCommand,
}

// Map unique IDs of components to their respected handler.
var ComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"minesweeperflagbutton": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Get user ID from the interaction.
		userID, _ := getUserID(i)

		// Check if the user has a game open.
		game, ok := Games[userID]
		if !ok {
			// User does not have a game open, send an error message.
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

		// Check if the flag button is associated with the user's game.
		if game.FlagID != i.Message.ID {
			// Flag button does not belong to the user's game, send an error message.
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

		// Respond to the interaction with a deferred message update.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		// Toggle the flag status.
		game.FlagEnabled = !game.FlagEnabled

		// Create the new flag button.
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

		// Update the label and style of the flag button based on the flag status.
		if game.FlagEnabled {
			flagButton.Label = "ON"
			flagButton.Style = discordgo.SuccessButton
		}
		flagRow.Components = append(flagRow.Components, flagButton, endGameButton)

		// Edit the old flag message with the new button.
		if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    game.ChannelID,
			ID:         game.FlagID,
			Components: []discordgo.MessageComponent{flagRow},
		}); err != nil {
			cmdError(s, i, err)
			return
		}
	},
	"endgamebutton": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		// Respond to the interaction with a deferred message update.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		// Get user ID from the interaction.
		userID, _ := getUserID(i)

		// Check if the user has a game open.
		game, ok := Games[userID]
		if !ok {
			// User does not have a game open, send an error message.
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

		// Handle the end of the game.
		HandleGameEnd(s, game, minesweeper.ManualEnd, false)
	},
	"profileleft": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		userID, _ := getUserID(i)
		// Respond to the interaction with a deferred message update.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		embed := i.Message.Embeds[0]

		ids := strings.Split(embed.Footer.Text, " - ")
		caller := ids[0]
		target := ids[1]

		var userString string
		var userImage string
		user, err := getUser(target, false)
		if err != nil {
			fmt.Println(err)
			userString = target
			userImage = "https://cdn.discordapp.com/embed/avatars/0.png"
		}
		if err == nil {
			userString = user.Username
			userImage = user.AvatarURL("")
		}

		page, err := strconv.Atoi(embed.Description[6:])
		if err != nil {
			cmdError(s, i, err)
			return
		}

		page--

		if userID != caller {
			return
		}

		userData := getUserData(target)

		title := fmt.Sprintf("%s's Achievements", userString)
		fields, components := getFieldsAndComponents(userData, page)
		desc := fmt.Sprintf("Page #%d", page+1)

		newEmbed := discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       title,
			Description: desc,
			Color:       randomEmbedColor(),
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s - %s", caller, target),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: userImage,
			},
		}

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				&newEmbed,
			},
			Components: &components,
		}); err != nil {
			fmt.Println(err)
		}
	},
	"profileright": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		userID, _ := getUserID(i)
		// Respond to the interaction with a deferred message update.
		go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
		embed := i.Message.Embeds[0]

		ids := strings.Split(embed.Footer.Text, " - ")
		caller := ids[0]
		target := ids[1]

		var userString string
		var userImage string
		user, err := getUser(target, false)
		if err != nil {
			fmt.Println(err)
			userString = target
			userImage = "https://cdn.discordapp.com/embed/avatars/0.png"
		}
		if err == nil {
			userString = user.Username
			userImage = user.AvatarURL("")
		}

		page, err := strconv.Atoi(embed.Description[6:])
		if err != nil {
			cmdError(s, i, err)
			return
		}

		if userID != caller {
			return
		}

		userData := getUserData(target)

		title := fmt.Sprintf("%s's Achievements", userString)
		fields, components := getFieldsAndComponents(userData, page)
		desc := fmt.Sprintf("Page #%d", page+1)

		newEmbed := discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       title,
			Description: desc,
			Color:       randomEmbedColor(),
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("%s - %s", caller, target),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: userImage,
			},
		}

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				&newEmbed,
			},
			Components: &components,
		}); err != nil {
			fmt.Println(err)
		}
	},
}

// Handle user interactions to the minesweeper board.
func HandleBoard(s *discordgo.Session, i *discordgo.InteractionCreate, positionx, positiony int) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	// Retrieve the user ID and check if a game exists for the user
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

	// Check if the game's board ID matches the ID of the triggering message
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

	// Set the start time if it hasn't been set yet
	if game.StartTime.IsZero() {
		game.StartTime = time.Now()
	}

	// Find the spot on the game board based on the provided coordinates
	spot := game.Game.FindSpot(positionx, positiony)

	for id, achievement := range AwardAchievements(game, minesweeper.Nothing, spot, false, false, true) {
		game.Achievements[id] = achievement
	}

	// Perform the appropriate action based on the FlagEnabled flag
	switch game.FlagEnabled {
	case true:
		var (
			event int
			chord bool
		)
		if spot.DisplayedType == minesweeper.Normal {
			event = game.Game.ChordSpot(spot)
		}
		if event != minesweeper.Nothing {
			// Handle the game end and respond with a deferred message update
			HandleGameEnd(s, game, event, true)
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			for id, achievement := range AwardAchievements(game, event, spot, chord, false, false) {
				game.Achievements[id] = achievement
			}
			return
		}
		if spot.DisplayedType == minesweeper.StartHere {
			game.Game.VisitSpot(spot)
			break
		}
		game.Game.FlagSpot(spot)
		for id, achievement := range AwardAchievements(game, event, spot, chord, true, false) {
			game.Achievements[id] = achievement
		}

	case false:
		var (
			gameEnd bool
			chord   bool
			event   int
		)
		if spot.DisplayedType == minesweeper.Normal {
			event = game.Game.ChordSpot(spot)
			chord = true
		}
		if event != minesweeper.Nothing {
			for id, achievement := range AwardAchievements(game, event, spot, chord, false, false) {
				game.Achievements[id] = achievement
			}
			// Handle the game end and respond with a deferred message update
			HandleGameEnd(s, game, event, true)
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			return
		}
		// Visit the spot and check if the game ends
		gameEnd, event = game.Game.VisitSpot(spot)
		for id, achievement := range AwardAchievements(game, event, spot, chord, false, false) {
			game.Achievements[id] = achievement
		}
		if gameEnd {
			// Handle the game end and respond with a deferred message update
			HandleGameEnd(s, game, event, true)
			go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			return
		}
	}

	// Update the game board message with the new content and components
	content := fmt.Sprintf("Here you go >~<\nTotal bombs: **%d**", game.Game.TotalBombs)
	board := GenerateBoard(game, false, false)
	editMessage := &discordgo.MessageEdit{
		Channel:    game.ChannelID,
		ID:         game.BoardID,
		Content:    &content,
		Components: board,
	}

	// Edit the message and handle any errors
	if _, err := s.ChannelMessageEditComplex(editMessage); err != nil {
		cmdError(s, i, err)
		return
	}

	// Respond with a deferred message update
	go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

// mapOptions maps the options of an interaction to a map for easier access
func mapOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}

// getUserID retrieves the user ID from an interaction and determines if it's in a guild
func getUserID(i *discordgo.InteractionCreate) (string, bool) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
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

// cmdError handles and logs command errors and responds to the interaction with an error message
func cmdError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	if err == nil {
		return
	}

	fmt.Println(err)

	go s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("An error occurred!\n```%s```", err.Error()),
		},
	})
}
