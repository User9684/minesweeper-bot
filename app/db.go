package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LeaderboardEntry struct {
	UserID string  `bson:"userId"`
	Time   float64 `bson:"time"`
	Spot   int     `bson:"spot"`
}

type Leaderboards struct {
	Easy   []LeaderboardEntry `bson:"easy"`
	Medium []LeaderboardEntry `bson:"Medium"`
	Hard   []LeaderboardEntry `bson:"hard"`
}

type GuildData struct {
	GuildID     string       `bson:"guildID"`
	Leaderboard Leaderboards `bson:"timeLeaderboard"`
}
type Blacklist struct {
	UserID  string `bson:"userID"`
	Message string `bson:"blacklistMessage"`
}
type LeaderboardMessage struct {
	GuildID    string `bson:"guildID"`
	ChannelID  string `bson:"channelID"`
	MessageID  string `bson:"messageID"`
	Difficulty int    `bson:"difficulty"`
}

var Collections = []string{
	"guilddata",
	"blacklists",
	"leaderboardmessages",
}

func DbInit() *mongo.Client {
	uri := os.Getenv("MONGOURI")
	fmt.Printf("Mongo URI set: \"%s\"\n", uri)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	return client
}

func CollectionCheck(d *mongo.Database) {
	collectionNames, err := d.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		fmt.Printf("Failed to get collection names %s\n", err)
	}

	for _, collectionName := range Collections {
		if !isInArray(collectionName, collectionNames) {
			fmt.Printf("Created collection %s\n", collectionName)
			d.CreateCollection(context.TODO(), collectionName)
		}
	}
}

func blacklistUser(userID, message string) {
	filter := bson.D{{
		Key:   "userID",
		Value: userID,
	}}
	newBlacklist := Blacklist{
		UserID:  userID,
		Message: message,
	}
	data, err := bson.Marshal(newBlacklist)
	if err != nil {
		fmt.Println(err)
		return
	}

	var update bson.M
	if err := bson.Unmarshal(data, &update); err != nil {
		return
	}

	request := d.Collection("blacklists").FindOneAndUpdate(
		context.TODO(),
		filter,
		bson.D{{
			Key:   "$set",
			Value: update,
		}},
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if err := request.Decode(&newBlacklist); err != nil {
		fmt.Println(err)
	}

	return
}

func unblacklistUser(userID string) {
	filter := bson.D{{
		Key:   "userID",
		Value: userID,
	}}

	request := d.Collection("blacklists").FindOneAndDelete(
		context.TODO(),
		filter,
		options.FindOneAndDelete(),
	)

	if err := request.Decode(&filter); err != nil {
		fmt.Println(err)
	}
}

func getGuildData(guildID string) GuildData {
	var guildData GuildData
	filter := bson.D{{
		Key:   "guildID",
		Value: guildID,
	}}
	d.Collection("guilddata").FindOne(context.TODO(), filter).Decode(&guildData)
	guildData.GuildID = guildID

	return guildData
}

func getBlacklist(userID string) Blacklist {
	var blacklistInfo Blacklist
	filter := bson.D{{
		Key:   "userID",
		Value: userID,
	}}
	d.Collection("blacklists").FindOne(context.TODO(), filter).Decode(&blacklistInfo)

	return blacklistInfo
}

func getLeaderboardMessages() []LeaderboardMessage {
	var results []LeaderboardMessage
	cursor, err := d.Collection("leaderboardmessages").Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println(err)
		return results
	}
	err = cursor.All(context.TODO(), &results)
	if err != nil {
		fmt.Println(err)
		return results
	}

	return results
}

func addLeaderboardMessage(guildID, channelID, messageID string, difficulty int) {
	filter := bson.D{{
		Key:   "guildID",
		Value: guildID,
	}, {
		Key:   "channelID",
		Value: channelID,
	}, {
		Key:   "messageID",
		Value: messageID,
	}}

	newData := LeaderboardMessage{
		GuildID:    guildID,
		ChannelID:  channelID,
		MessageID:  messageID,
		Difficulty: difficulty,
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

	request := d.Collection("leaderboardmessages").FindOneAndReplace(
		context.TODO(),
		filter,
		update,
		options.FindOneAndReplace().SetUpsert(true),
	)

	if err := request.Decode(&update); err != nil {
		fmt.Println(err)
	}
}

func removeLeaderboardMessage(messageID string) {
	filter := bson.D{{
		Key:   "messageID",
		Value: messageID,
	}}

	request := d.Collection("leaderboardmessages").FindOneAndDelete(
		context.TODO(),
		filter,
		options.FindOneAndDelete(),
	)

	if err := request.Decode(&filter); err != nil {
		fmt.Println(err)
	}
}
