package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LearderboardEntry struct {
	UserID string `bson:"userId"`
	Time   int    `bson:"time"`
}

type GuildData struct {
	ID          primitive.ObjectID  `bson:"_id"`
	GuildID     string              `bson:"guildid"`
	Leaderboard []LearderboardEntry `bson:"timeLeaderboard"`
}

var Collections = []string{
	"guilddata",
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

func getGuildData(guildID string) GuildData {
	var guildData GuildData
	filter := bson.D{{
		Key:   "guildid",
		Value: guildID,
	}}
	d.Collection("guilddata").FindOne(context.TODO(), filter).Decode(&guildData)

	return guildData
}
