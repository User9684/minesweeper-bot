package main

import (
	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "minesweeper",
		Description: "Generate a 5x5 minesweeper board",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "difficulty",
				Description: "Difficulty level",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Easy",
						Value: "easy",
					},
					{
						Name:  "Medium",
						Value: "medium",
					},
					{
						Name:  "Hard",
						Value: "hard",
					},
				},
			},
		},
	},
}
