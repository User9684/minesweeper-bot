package main

import (
	"context"
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var autoEditChannel chan struct{}

func orderBySpot(entries []LeaderboardEntry) []LeaderboardEntry {
	orderedLeaderboard := make([]LeaderboardEntry, len(entries))

	for _, entry := range entries {
		if entry.UserID == "" {
			continue
		}
		orderedLeaderboard[entry.Spot] = entry
	}

	return orderedLeaderboard
}

func getLeaderboard(guildID string, difficulty int) []LeaderboardEntry {
	guildData := getGuildData(guildID)
	leaderboards := guildData.Leaderboard
	var leaderboard []LeaderboardEntry

	switch difficulty {
	case minesweeper.Easy:
		leaderboard = leaderboards.Easy
	case minesweeper.Medium:
		leaderboard = leaderboards.Medium
	case minesweeper.Hard:
		leaderboard = leaderboards.Hard
	}

	orderedLeaderboard := orderBySpot(leaderboard)

	return orderedLeaderboard
}

func addToLeaderboard(guildID string, difficulty int, newEntry LeaderboardEntry) {
	currentLeaderboard := getLeaderboard(guildID, difficulty)
	var dontReorder bool
	// Remove duplicate ID if new is shorter in length.
	for index, leaderboardEntry := range currentLeaderboard {
		if leaderboardEntry.UserID != newEntry.UserID {
			continue
		}
		if leaderboardEntry.Time < newEntry.Time {
			// Return because the new entry is longer than the past one.
			return
		}
		if leaderboardEntry.Spot == 0 {
			newEntry.Spot = 0
			currentLeaderboard[index] = newEntry
			dontReorder = true
			break
		}
		currentLeaderboard = append(currentLeaderboard[:index], currentLeaderboard[index+1:]...)
	}

	if !dontReorder {
		newEntry.Spot = len(currentLeaderboard)
		fmt.Println(newEntry.Spot)

		for i, leaderboardEntry := range currentLeaderboard {
			if leaderboardEntry.Time > newEntry.Time {
				newEntry.Spot--
				currentLeaderboard[i].Spot++
				continue
			}
			break
		}

		currentLeaderboard = append(currentLeaderboard, newEntry)
	}

	for _, leaderboardEntry := range currentLeaderboard {
		fmt.Printf("%s %d\n", leaderboardEntry.UserID, leaderboardEntry.Spot)
	}

	currentLeaderboard = orderBySpot(currentLeaderboard)

	if len(currentLeaderboard) > 10 {
		currentLeaderboard = currentLeaderboard[:10]
	}

	filter := bson.D{{
		Key:   "guildID",
		Value: guildID,
	}}

	newData := getGuildData(guildID)

	switch difficulty {
	case minesweeper.Easy:
		newData.Leaderboard.Easy = currentLeaderboard
	case minesweeper.Medium:
		newData.Leaderboard.Medium = currentLeaderboard
	case minesweeper.Hard:
		newData.Leaderboard.Hard = currentLeaderboard
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

	request := d.Collection("guilddata").FindOneAndUpdate(
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
}

func generateLeaderboardEmbed(guildID, guildName, difficultyString string) (discordgo.MessageEmbed, error) {
	var difficulty int
	switch difficultyString {
	case "easy":
		difficulty = minesweeper.Easy
	case "medium":
		difficulty = minesweeper.Medium
	case "hard":
		difficulty = minesweeper.Hard
	}

	leaderboard := getLeaderboard(guildID, difficulty)

	r := rand.Intn(255)
	g := rand.Intn(255)
	b := rand.Intn(255)
	colorInt := (r << 16) + (g << 8) + b

	embed := discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       fmt.Sprintf("%s Leaderboard", guildName),
		Description: fmt.Sprintf("Leaderboard for **%s** mode", strings.ToUpper(difficultyString)),
		Color:       colorInt,
	}

	for _, entry := range leaderboard {
		userString := entry.UserID
		user, err := s.User(entry.UserID, RequestOption)
		if err == nil {
			userString = user.Username
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%vs", entry.Time))
		if err != nil {
			fmt.Println(err)
			return discordgo.MessageEmbed{}, err
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   userString,
			Value:  humanizetime.HumanizeDuration(duration, 3),
			Inline: false,
		})
	}

	return embed, nil
}

func editConfiguredMessages() {
	fmt.Println("Editing leaderboard messages...")
	messages := getLeaderboardMessages()

	for _, message := range messages {
		guild, err := s.Guild(message.GuildID)
		if err != nil {
			removeLeaderboardMessage(message.MessageID)
			fmt.Println(err)
			continue
		}

		var difficultyString string
		switch message.Difficulty {
		case minesweeper.Easy:
			difficultyString = "easy"
		case minesweeper.Medium:
			difficultyString = "medium"
		case minesweeper.Hard:
			difficultyString = "hard"
		}

		embed, err := generateLeaderboardEmbed(guild.ID, guild.Name, difficultyString)
		if err != nil {
			removeLeaderboardMessage(message.MessageID)
			fmt.Println(err)
			continue
		}

		if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:      message.MessageID,
			Channel: message.ChannelID,
			Embeds:  []*discordgo.MessageEmbed{&embed},
		}); err != nil {
			removeLeaderboardMessage(message.MessageID)
			fmt.Println(err)
			continue
		}
	}
}

func startAutoEdit() {
	ticker := time.NewTicker(5 * time.Minute)
	autoEditChannel = make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				editConfiguredMessages()
			case <-autoEditChannel:
				ticker.Stop()
				return
			}
		}
	}()
}
