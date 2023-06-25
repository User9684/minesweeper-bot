package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Gets the ping of the bot to Discord",
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
	{
		Name:        "leaderboard",
		Description: "Gets the leaderboard for the current server, or global.",
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
			{
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Name:        "global",
				Description: "Get the global leaderboard",
				Required:    false,
			},
		},
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
					{
						Name:        "difficulty",
						Description: "Difficulty to use for leaderboard",
						Type:        discordgo.ApplicationCommandOptionString,
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
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "win",
				Description: "Force a minesweeper win for a user",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "winner winner chicken dinner",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "reveal",
				Description: "Reveal all the spots for a user's game",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "user",
						Description: "user's game to view",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "restartticker",
				Description: "Restart the leaderboard editing ticker",
			},
		},
	},
}

// Register commands using the Commands slice.
func RegisterCommands(s *discordgo.Session) {
	registeredCommands := make([]*discordgo.ApplicationCommand, len(Commands))

	for i, v := range Commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			fmt.Printf("Cannot create '%v' command\n%v\n", v.Name, err)
		}

		registeredCommands[i] = cmd
	}

	fmt.Printf("Registered %d commands\n", len(registeredCommands))
}
