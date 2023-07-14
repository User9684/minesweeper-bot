package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type cachedUser struct {
	User       *discordgo.User
	LastAccess int64
	StopTimer  chan struct{}
}

// Cache userIDs to user objects.
var userCache = make(map[string]*cachedUser)

// Gets user from cache if present, else fetch from API.
func getUser(userid string, recache bool) (user *discordgo.User, err error) {
	cachedUserData, ok := userCache[userid]
	if ok && !recache {
		return cachedUserData.User, nil
	}
	if ok && cachedUserData.LastAccess <= 5*60 {
		return cachedUserData.User, nil
	}

	user, err = s.User(userid, RequestOption)
	if err != nil {
		return
	}

	channel := make(chan struct{})
	timer := time.NewTimer(time.Duration(15 * time.Minute))
	go func() {
		for {
			select {
			case <-timer.C:
				delete(userCache, userid)
				return
			case <-channel:
				timer.Stop()
				return
			}
		}
	}()
	userCache[userid] = &cachedUser{
		User:       user,
		LastAccess: time.Now().Unix(),
		StopTimer:  channel,
	}

	return
}

// RegisterEvents registers the event handlers.
func RegisterEvents() {
	// Ready event.
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		intents := s.Identify.Intents

		fmt.Printf("Logged in as: %v#%v\nIntents: %v\n", s.State.User.Username, s.State.User.Discriminator, intents)

		// Set bot status accordingly to bot config.
		config := getBotConfig()
		if config.Presence.Status != "" {
			setBotPresence(config.Presence)
		}
	})

	// Interaction handler.
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()
		var ignoreBlacklist bool

		if i.Type == discordgo.InteractionApplicationCommand &&
			i.ApplicationCommandData().Name == "admin" {
			ignoreBlacklist = true
		}

		userID, _ := getUserID(i)

		blacklistData := getBlacklist(userID)
		if !ignoreBlacklist && blacklistData.Message != "" {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   1 << 6,
					Content: fmt.Sprintf("You have been blacklisted.\nMessage given:\n%s", blacklistData.Message),
				},
			}); err != nil {
				cmdError(s, i, err)
			}
			return
		}

		getUser(userID, false)

		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			HandleCommand(s, i)
		case discordgo.InteractionMessageComponent:
			HandleComponent(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, i *discordgo.UserUpdate) {
		defer func() {
			if err := recover(); err != nil {
				handlePanic(err)
			}
		}()

		if _, err := getUser(i.User.ID, true); err != nil {
			fmt.Println(err)
		}
	})
}

// Handle application command interactions.
func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	userid, _ := getUserID(i)
	commandData := i.ApplicationCommandData()
	if handler, ok := CommandHandlers[commandData.Name]; ok {
		go handler(s, i)
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   1 << 6,
			Content: "Invalid command! Deleting...",
		},
	}); err != nil {
		cmdError(s, i, err)
	}

	commandID := i.Interaction.ApplicationCommandData().ID

	fmt.Printf("Invalid command detected.\nCommand ID: %s\nCommand used by: %s\n", commandID, userid)
	s.ApplicationCommandDelete(s.State.User.ID, "", commandID)
}

// Handle message component interactions.
func HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
	customID := i.MessageComponentData().CustomID

	boardPositionMatches := BoardPositionRegex.FindAllStringSubmatch(customID, -1)

	if len(boardPositionMatches) > 0 {
		match := boardPositionMatches[0]
		x, err := strconv.Atoi(match[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		y, err := strconv.Atoi(match[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		HandleBoard(s, i, x, y)
		return
	}

	if handler, ok := ComponentHandlers[customID]; ok {
		go handler(s, i)
		return
	}

	fmt.Printf("Invalid component interaction detected.\nCustom ID: %s\nCommand used by: %s\n", customID, i.Interaction.Member.User.ID)
}

func setBotPresence(config PresenceData) {
	activity := &discordgo.Activity{
		Name: config.Status,
	}

	switch config.Presence {
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

		streamURL := config.URL
		if streamURL == "" {
			streamURL = "https://www.youtube.com/watch?v=Pr2ONUSGMgQ"
		}

		activity.URL = streamURL
	case "CUSTOM":
		activity.Type = discordgo.ActivityTypeCustom
	}

	if err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{activity},
	}); err != nil {
		fmt.Printf("Failed to set bot presence!\n%v\n", err)
	}
}
