package main

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"main/minesweeper"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

type MinesweeperGame struct {
	GuildID     string
	ChannelID   string
	BoardID     string
	FlagID      string
	UserID      string
	Difficulty  string
	FlagEnabled bool
	Won         bool
	StartTime   time.Time
	Game        *minesweeper.Game
	EndGameChan *chan struct{}
}

var s *discordgo.Session
var c *mongo.Client
var d *mongo.Database
var Games = make(map[string]*MinesweeperGame)
var ChannelMutex = make(map[string]*sync.Mutex)
var BoardPositionRegex = regexp.MustCompile(`boardx(\d+)y(\d+)`)
var MessageLinkRegex = regexp.MustCompile(`(?:http(?:s)?://)(?:(?:canary|ptb).)?discord.com/channels/(\d+|@me)/(\d+)/(\d+)`)
var UserStatsFormatString = "**Wins:** %d\n**Losses:** %d\n**Personal Best:** %s\n**Personal Worst:** %s"
var Admins = make(map[string]bool)
var EndAfter int64

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
	// Load variables from .env if it exists.
	godotenv.Load()
	// Mongo setup.
	fmt.Println("Connecting to mongo...")
	c = DbInit()
	d = c.Database("minesweeper")
	fmt.Println("Verifying all collections...")
	CollectionCheck(d)
	// Config setup.
	fmt.Println("Setting up admin map...")
	for _, userID := range strings.Split(os.Getenv("ADMINS"), " ") {
		Admins[userID] = true
		fmt.Printf("Added %s to admin map.\n", userID)
	}
	fmt.Println("Fetching auto-end setting")
	converted, err := strconv.Atoi(os.Getenv("END_GAME_AFTER"))
	EndAfter = int64(converted)
	if err != nil {
		fmt.Printf("Failed to setup auto-end\n%v\n", err)
		EndAfter = int64(0)
	}

	// Bot setup.
	fmt.Println("Starting the bot...")
	BotInit()

	err = s.Open()
	if err != nil {
		fmt.Printf("Cannot open the session\n%v\n", err)
		return
	}

	RegisterCommands(s)

	fmt.Println("Starting leaderboard edit ticker...")
	startAutoEdit()

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

func handlePanic(err interface{}) {
	stackSplit := strings.Split(string(debug.Stack()), "\n")
	stackTrace := strings.Join(append(stackSplit[:1], stackSplit[7:]...), "\n")
	log := fmt.Sprintf("Recovered from panic\n%v\n%s", err, stackTrace)

	fmt.Println(log)

	panicLog := os.Getenv("PANIC_CHANNEL")
	panicMessage := os.Getenv("PANIC_MESSAGE")

	if panicMessage == "" {
		panicMessage = "Recovered from a panic!"
	}

	if panicLog == "" {
		fmt.Println("No panic log channel set")
		return
	}

	if _, err := s.ChannelMessageSendComplex(panicLog, &discordgo.MessageSend{
		Content: panicMessage,
		Files: []*discordgo.File{
			{
				Name:        "log.txt",
				ContentType: "attachment",
				Reader:      strings.NewReader(log),
			},
		},
	}, RequestOption); err != nil {
		fmt.Println(err)
	}
}
