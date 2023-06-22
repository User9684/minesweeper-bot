package main

import (
	"context"
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
	// Add to global leaderboard.
	addToLeaderboard("global", difficulty, newEntry)

	currentLeaderboard := getLeaderboard(guildID, difficulty)
	// Remove duplicate ID if new is shorter in length.
	for index, leaderboardEntry := range currentLeaderboard {
		if leaderboardEntry.UserID != newEntry.UserID {
			continue
		}
		if leaderboardEntry.Time < newEntry.Time {
			// Return because the new entry is longer than the past one.
			return
		}
		currentLeaderboard = append(currentLeaderboard[:index], currentLeaderboard[index+1:]...)
	}

	newEntry.Spot = len(currentLeaderboard)

	for i, leaderboardEntry := range currentLeaderboard {
		if leaderboardEntry.Time > newEntry.Time {
			newEntry.Spot--
			currentLeaderboard[i].Spot++
			continue
		}
		break
	}

	currentLeaderboard = append(currentLeaderboard, newEntry)
	currentLeaderboard = orderBySpot(currentLeaderboard)[:10]

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
