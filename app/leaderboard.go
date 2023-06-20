package main

import (
	"context"
	"encoding/json"
	"fmt"
	"main/minesweeper"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
	// Remove duplicate ID if new is shorter in length.
	for index, leaderboardEntry := range currentLeaderboard {
		if leaderboardEntry.UserID != newEntry.UserID {
			continue
		}
		if leaderboardEntry.Time < newEntry.Time {
			continue
		}
		currentLeaderboard = append(currentLeaderboard[:index], currentLeaderboard[index+1:]...)
	}

	for _, leaderboardEntry := range currentLeaderboard {
		if leaderboardEntry.Time > newEntry.Time {
			newEntry.Spot = leaderboardEntry.Spot
			leaderboardEntry.Spot++
			continue
		}
		newEntry.Spot++
		break
	}

	currentLeaderboard = orderBySpot(append(currentLeaderboard, newEntry))

	filter := bson.D{{
		Key:   "guildID",
		Value: guildID,
	}}

	newData := GuildData{
		GuildID: guildID,
	}

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

	d.Collection("guilddata").
		FindOneAndUpdate(
			context.TODO(),
			filter,
			bson.M{
				"$set": data,
			},
			options.FindOneAndUpdate().SetUpsert(true),
		)

	b, _ := json.MarshalIndent(currentLeaderboard, "", "	")
	fmt.Println(string(b))
}
