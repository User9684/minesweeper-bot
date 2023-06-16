package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"main/minesweeper"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
)

type MinesweeperGame struct {
	GuildID     string
	ChannelID   string
	BoardID     string
	FlagID      string
	UserID      string
	Difficulty  string
	flagEnabled bool
	Game        *minesweeper.Game
}

var s *discordgo.Session
var c *mongo.Client
var d *mongo.Database
var Games = make(map[string]*MinesweeperGame)
var BoardPositionRegex = regexp.MustCompile(`boardx(\d+)y(\d+)`)

func BotInit() {
	// Create bot client.
	session, err := discordgo.New(os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	// Register intents.
	session.Identify.Intents |= discordgo.IntentGuilds
	session.Identify.Intents |= discordgo.IntentGuildMembers

	// Setup state.
	session.StateEnabled = true
	session.State.TrackChannels = true
	session.State.TrackMembers = true

	s = session

	RegisterEvents()
}

func main() {
	// Mongo setup.
	fmt.Println("Connecting to mongo...")
	c = DbInit()
	d = c.Database("ABL")
	fmt.Println("Verifying all collections...")
	CollectionCheck(d)

	// Bot setup.
	fmt.Println("Starting the bot...")
	BotInit()

	err := s.Open()
	if err != nil {
		fmt.Printf("Cannot open the session: %v\n", err)
		return
	}

	CmdInit(s)

	// Waits for SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func isInArray(value string, array []string) bool {
	for _, v := range array {
		if value == v {
			return true
		}
	}

	return false
}
