package main

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func RegisterEvents() {
	// Ready event.
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		intents := s.Identify.Intents

		fmt.Printf("Logged in as: %v#%v\nIntents: %v\n", s.State.User.Username, s.State.User.Discriminator, intents)
	})

	// Command handler.
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			HandleCommand(s, i)
		case discordgo.InteractionMessageComponent:
			HandleComponent(s, i)
		}
	})

}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := InteractionHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Invalid command! Deleting...",
		},
	})

	commandID := i.Interaction.ApplicationCommandData().ID

	fmt.Printf("Invalid command detected.\nCommand ID: %s\nCommand used by: %s\n", commandID, i.Interaction.Member.User.ID)
	s.ApplicationCommandDelete(s.State.User.ID, "", commandID)
}

func HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	if h, ok := InteractionHandlers[customID]; ok {
		h(s, i)
		return
	}

	fmt.Printf("Invalid component interaction detected.\nCustom ID: %s\nCommand used by: %s\n", customID, i.Interaction.Member.User.ID)
}
