package main

import (
	"context"
	"fmt"
	"main/minesweeper"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// The admin command is so long that I'm moving it to it's own dedicated file.

func AdminCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var ignoreRecover bool
	defer func() {
		if ignoreRecover {
			return
		}
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	userID, _ := getUserID(i)

	// Check if the user is an admin.
	if a, ok := Admins[userID]; !a || !ok {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: "No.",
			},
		}); err != nil {
			cmdError(s, i, err)
		}

		return
	}

	subcommand := i.Interaction.ApplicationCommandData().Options[0]
	optionMap := mapOptions(subcommand.Options)

	switch subcommand.Name {
	case "blacklist":
		target := optionMap["target"].UserValue(s).ID
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
			cmdError(s, i, err)
		}

	case "unblacklist":
		target := optionMap["target"].UserValue(s).ID

		unblacklistUser(target)

		replyContent := fmt.Sprintf("Removed blacklist for `%s`", target)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: replyContent,
			},
		}); err != nil {
			cmdError(s, i, err)
		}

	case "leaderboardmsg":
		// Respond with a deferred message update initially.
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: 1 << 6,
			},
		})
		match := MessageLinkRegex.FindStringSubmatch(optionMap["message"].StringValue())

		if match[1] == "@me" {
			content := "Automatic leaderboard editing not supported in DMs!"
			if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			}); err != nil {
				cmdError(s, i, err)
			}
			return
		}

		var difficulty int
		switch optionMap["difficulty"].Value {
		case "easy":
			difficulty = minesweeper.Easy
		case "medium":
			difficulty = minesweeper.Medium
		case "hard":
			difficulty = minesweeper.Hard
		}

		addLeaderboardMessage(match[1], match[2], match[3], difficulty)

		content := fmt.Sprintf("Added %s to automatic editing for difficulty **%s**!", optionMap["message"].StringValue(), optionMap["difficulty"].StringValue())
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		}); err != nil {
			cmdError(s, i, err)
		}

	case "win":
		target := optionMap["target"].UserValue(s).ID
		game, ok := Games[target]
		// Check if the user has a game open.
		if !ok {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: "That user doesn't have a game open!",
				},
			}); err != nil {
				cmdError(s, i, err)
			}
			return
		}

		replyContent := fmt.Sprintf("Forcewon `%s`'s game", target)

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: replyContent,
			},
		}); err != nil {
			cmdError(s, i, err)
		}

		HandleGameEnd(s, game, minesweeper.Won, false)

	case "reveal":
		target := optionMap["target"].UserValue(s).ID
		game, ok := Games[target]
		// Check if the user has a game open.
		if !ok {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: "That user doesn't have a game open!",
				},
			}); err != nil {
				cmdError(s, i, err)
			}
			return
		}

		board := GenerateBoard(game, false, true)
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:      1 << 6,
				Content:    "👁️",
				Components: board,
			},
		}); err != nil {
			cmdError(s, i, err)
		}

	case "restartticker":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})

		close(autoEditChannel)
		editConfiguredMessages()
		startAutoEdit()

		content := "Restarted ticker."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

	case "presence":
		str := optionMap["status"].StringValue()
		pre := optionMap["presence"].StringValue()

		// Update saved status.
		filter := bson.D{{
			Key:   "botID",
			Value: s.State.User.ID,
		}}

		newData := getBotConfig()
		newData.BotID = s.State.User.ID
		newData.Presence = PresenceData{
			Presence: pre,
			Status:   str,
		}

		if pre == "CLEAR" {
			newData.Presence = PresenceData{}
		}

		data, err := bson.Marshal(newData)
		if err != nil {
			fmt.Println(err)
			return
		}

		var update bson.M
		if err := bson.Unmarshal(data, &update); err != nil {
			return
		}

		request := d.Collection("botconfig").FindOneAndUpdate(
			context.TODO(),
			filter,
			bson.D{{
				Key:   "$set",
				Value: update,
			}},
			options.FindOneAndUpdate().SetUpsert(true),
		)

		if err := request.Decode(&newData); err != nil {
			fmt.Println(err)
		}

		// Update the actual bot status
		activity := &discordgo.Activity{
			Name: str,
		}

		switch pre {
		case "WATCHING":
			activity.Type = discordgo.ActivityTypeWatching
		case "PLAYING":
			activity.Type = discordgo.ActivityTypeGame
		case "LISTENING":
			activity.Type = discordgo.ActivityTypeListening
		case "COMPETING":
			activity.Type = discordgo.ActivityTypeCompeting
		case "STREAMING":
			activity.Type = discordgo.ActivityTypeStreaming

			var streamURL string
			stval, ok := optionMap["streaming"]
			if ok {
				streamURL = stval.StringValue()
			}
			if !ok {
				streamURL = "https://www.youtube.com/watch?v=Pr2ONUSGMgQ"
			}

			activity.URL = streamURL
		case "CLEAR":
			if err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
				Activities: []*discordgo.Activity{},
			}); err != nil {
				cmdError(s, i, err)
			}
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: "Cleared bot presence!",
				},
			}); err != nil {
				cmdError(s, i, err)
			}
			return
		}

		if err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{activity},
		}); err != nil {
			cmdError(s, i, err)
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: fmt.Sprintf("Set the bot's %s presence to %s", pre, str),
			},
		}); err != nil {
			cmdError(s, i, err)
		}
	case "panic":
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   1 << 6,
				Content: "Causing a PANIC...",
			},
		}); err != nil {
			cmdError(s, i, err)
			return
		}

		if !optionMap["recover"].BoolValue() {
			ignoreRecover = true
		}

		smallSlice := make([]int, 2)
		_ = smallSlice[10]
	}
}
