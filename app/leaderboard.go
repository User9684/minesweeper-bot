package main

import (
	"context"
	"fmt"
	"main/humanizetime"
	"main/minesweeper"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var autoEditChannel chan struct{}

func orderBySpot(entries []LeaderboardEntry) []LeaderboardEntry {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
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
	defer func() {
		if err := recover(); err != nil {
			handlePanic(err)
		}
	}()
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
		currentLeaderboard = append(currentLeaderboard, newEntry)
		sort.Slice(currentLeaderboard, func(i, j int) bool {
			return currentLeaderboard[i].Time < currentLeaderboard[j].Time
		})

		for spot := range currentLeaderboard {
			entry := currentLeaderboard[spot]
			entry.Spot = spot
			currentLeaderboard[spot] = entry
		}
	}

	currentLeaderboard = orderBySpot(currentLeaderboard)

	if len(currentLeaderboard) > 10 {
		currentLeaderboard = currentLeaderboard[:10]
	}

	newData := getGuildData(guildID)

	switch difficulty {
	case minesweeper.Easy:
		newData.Leaderboard.Easy = currentLeaderboard
	case minesweeper.Medium:
		newData.Leaderboard.Medium = currentLeaderboard
	case minesweeper.Hard:
		newData.Leaderboard.Hard = currentLeaderboard
	}

	filter := bson.D{{
		Key:   "guildID",
		Value: guildID,
	}}

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

	embed := discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       fmt.Sprintf("%s Leaderboard", guildName),
		Description: fmt.Sprintf("Leaderboard for **%s** mode", strings.ToUpper(difficultyString)),
		Color:       randomEmbedColor(),
	}

	for _, entry := range leaderboard {
		userString := entry.UserID
		user, err := s.User(entry.UserID, RequestOption)
		if err == nil {
			userString = user.Username
		}

		duration, err := time.ParseDuration(fmt.Sprintf("%fs", entry.Time))
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
		guild, err := s.State.Guild(message.GuildID)
		if err != nil {
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
