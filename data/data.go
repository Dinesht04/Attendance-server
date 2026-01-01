package data

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectToMongo() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(fmt.Errorf("DB conn err: %w", err))
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		panic(fmt.Errorf("Ping Issue: %w", err))
	}

	return client
}
