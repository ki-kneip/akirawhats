package db

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
	once        sync.Once
)

func Connect(ctx context.Context) {
	once.Do(func() {
		uri := os.Getenv("MONGO_URI")
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}
		dbName := os.Getenv("MONGO_DB")
		if dbName == "" {
			dbName = "akirawhats"
		}

		client, err := mongo.Connect(options.Client().ApplyURI(uri))
		if err != nil {
			log.Fatalf("mongodb connect: %v", err)
		}

		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := client.Ping(pingCtx, nil); err != nil {
			log.Fatalf("mongodb ping: %v", err)
		}

		mongoClient = client
		mongoDB = client.Database(dbName)
		log.Printf("connected to mongodb: %s / %s", uri, dbName)
	})
}

func Disconnect(ctx context.Context) {
	if mongoClient != nil {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongodb disconnect: %v", err)
		}
	}
}

func Collection(name string) *mongo.Collection {
	return mongoDB.Collection(name)
}
