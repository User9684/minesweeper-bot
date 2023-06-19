package main

import (
	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Gets the ping of the bot to Discord",
	},
	{
		Name:        "admin",
		Description: "Admin commands xd",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "blacklist",
				Description: "blacklist a nerd",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "nerd to blacklist",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
					{
						Name:        "message",
						Description: "Message to display in blacklist response",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "unblacklist",
				Description: "unblacklist a nerd",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "nerd to unblacklist",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "leaderboardmsg",
				Description: "Add / remove an automatically edited leaderboard message",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "message",
						Description: "Message link of leaderboard",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
				},
			},
		},
	},
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
