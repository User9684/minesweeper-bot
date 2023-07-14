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
		Name:        "custom",
		Description: "Generate a custom minesweeper game",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "bombs",
				Description: "Bomb count",
				Type:        discordgo.ApplicationCommandOptionInteger,
				Required:    true,
			},
			{
				Name:        "surroundingbombs",
				Description: "Allow bombs to be placed surrounding the start spot. Ignored if nostartspot enabled.",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
			},
			{
				Name:        "nostartspot",
				Description: "Allow the user to start game in any spot",
				Type:        discordgo.ApplicationCommandOptionBoolean,
				Required:    true,
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
		Name:        "profile",
		Description: "Gets either your profile or a targets profile",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "target",
				Description: "User to get the profile of",
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
						Name:        "target",
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
						Name:        "target",
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
						Name:        "target",
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
						Name:        "target",
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
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "presence",
				Description: "Set the presence of the bot",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "status",
						Description: "Presence string",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
					},
					{
						Name:        "presence",
						Description: "Type of status",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "WATCHING",
								Value: "WATCHING",
							},
							{
								Name:  "PLAYING",
								Value: "PLAYING",
							},
							{
								Name:  "LISTENING",
								Value: "LISTENING",
							},
							{
								Name:  "COMPETING",
								Value: "COMPETING",
							},
							{
								Name:  "STREAMING",
								Value: "STREAMING",
							},
							{
								Name:  "CLEAR",
								Value: "CLEAR",
							},
						},
					},
					{
						Name:        "streaming",
						Description: "Link to use for streaming presence",
						Type:        discordgo.ApplicationCommandOptionString,
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "panic",
				Description: "Purposely cause a panic",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "recover",
						Description: "Recover the panic to prevent process kill",
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Required:    true,
					},
				},
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
