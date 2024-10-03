package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"blogr.moe/backend/logs"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var DB_Main *mongo.Database
var DB_Users *mongo.Database
var DB_UserList *mongo.Database

func init() {
	err := godotenv.Load()
	if err != nil {
		logs.Error("Error loading .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		logs.Error("Error creating MongoDB client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		logs.Error("Error connecting to MongoDB")
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		logs.Error("Error pinging MongoDB")
	}

	DB_Main = client.Database("blogr")
	DB_Users = client.Database("blogr_users")
	DB_UserList = client.Database("blogr_userlist")

	//create collections
	err = ensureStatsDocumentExists()
	if err != nil {
		logs.Error("Error ensuring stats document exists")
	}

}
func ensureStatsDocumentExists() error {
	ctx := context.Background()
	filter := bson.M{"_id": "stats"} // Unique identifier for the stats document
	update := bson.M{
		"$setOnInsert": bson.M{
			"post_count":    0,
			"user_count":    0,
			"comment_count": 0,
			"view_count":    0,
			"last_updated":  time.Now().Format(time.RFC3339),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := DB_Main.Collection("stats").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("error ensuring stats document exists: %v", err)
	}

	return nil
}
func GetCollection(collection string) *mongo.Collection {
	return DB_Main.Collection(collection)
}

func GetTotalPostCount() error {
	var total int
	ctx := context.Background()

	numcollections, err := DB_Users.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("error fetching post count: %v", err)
	}

	for _, collection := range numcollections {
		count, err := DB_Users.Collection(collection).EstimatedDocumentCount(ctx)
		if err != nil {
			return fmt.Errorf("error fetching post count: %v", err)
		}
		total += int(count)
	}

	// Update the stats document with the new post count
	_, err = DB_Main.Collection("stats").UpdateOne(
		ctx,
		bson.M{"_id": "stats"}, // Match the stats document
		bson.M{"$set": bson.M{"post_count": total}}, // Update the post count
	)
	if err != nil {
		return fmt.Errorf("error updating totalposts: %v", err)
	}

	return nil
}

func GetTotalUserCount() error {
	ctx := context.Background()
	count, err := DB_UserList.Collection("users").EstimatedDocumentCount(ctx)
	if err != nil {
		return fmt.Errorf("error fetching user count: %v", err)
	}

	// Update the stats document with the new user count
	_, err = DB_Main.Collection("stats").UpdateOne(
		ctx,
		bson.M{"_id": "stats"}, // Match the stats document
		bson.M{"$set": bson.M{"user_count": count}}, // Update the user count
	)
	if err != nil {
		return fmt.Errorf("error updating totalusers: %v", err)
	}

	return nil
}

func GetStats() (bson.M, error) {
	ctx := context.Background()
	var stats bson.M
	err := DB_Main.Collection("stats").FindOne(ctx, bson.M{"_id": "stats"}).Decode(&stats)
	if err != nil {
		return nil, fmt.Errorf("error fetching stats: %v", err)
	}

	return stats, nil
}
