package main

import (
	"fmt"
	"math/rand"
	"net/http"
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
	GuildID      string
	ChannelID    string
	BoardID      string
	FlagID       string
	UserID       string
	Difficulty   string
	Flags        int64
	StartTime    time.Time
	Achievements map[int]Achievement
	Game         *minesweeper.Game
	EndGameChan  *chan struct{}
}

var s *discordgo.Session
var c *mongo.Client
var d *mongo.Database
var Games = make(map[string]*MinesweeperGame)
var ChannelMutex = make(map[string]*sync.Mutex)
var BoardPositionRegex = regexp.MustCompile(`boardx(\d+)y(\d+)`)
var MessageLinkRegex = regexp.MustCompile(`(?:http(?:s)?://)(?:(?:canary|ptb).)?discord.com/channels/(\d+|@me)/(\d+)/(\d+)`)
var UserStatsFormatString = "**Wins:** %d\n**Losses:** %d\n**Winstreak:** %d\n**Personal Best:** %s\n**Personal Worst:** %s"
var Admins = make(map[string]bool)
var EndAfter int64
var TGGStatsURI string

func BotInit() {
	var token = os.Getenv("TOKEN")
	// Create bot client.
	session, err := discordgo.New(token)
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

	botID := strings.Split(token, ".")[0]
	TGGStatsURI = fmt.Sprintf("https://top.gg/api/bots/%s/stats", botID)

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

	if auth := os.Getenv("TOPGG_BOT_TOKEN"); auth != "" {
		fmt.Println("Starting top.gg auto stats uploader...")
		startTGGUpdater(auth)
	}

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

func isInIntArray(value int, array []int) bool {
	for _, v := range array {
		if value == v {
			return true
		}
	}

	return false
}

func randomEmbedColor() int {
	r := rand.Intn(255)
	g := rand.Intn(255)
	b := rand.Intn(255)
	return (r << 16) + (g << 8) + b
}

var tggChannel chan struct{}

func startTGGUpdater(auth string) {
	ticker := time.NewTicker(5 * time.Minute)
	tggChannel = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				data := fmt.Sprintf("{\"server_count\":%d}", len(s.State.Guilds))
				request, err := http.NewRequest(http.MethodPost, TGGStatsURI, strings.NewReader(data))
				if err != nil {
					fmt.Println(err)
					break
				}
				request.Header.Set("Authorization", auth)

				_, err = http.DefaultClient.Do(request)
				if err != nil {
					fmt.Println(err)
				}
			case <-tggChannel:
				ticker.Stop()
				return
			}
		}
	}()
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
