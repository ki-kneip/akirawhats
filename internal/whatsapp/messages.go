package whatsapp

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"kikneip.com/akirawhats/internal/db"
)

const messagesCollection = "messages"

const (
	MsgDirectionIn  = "in"
	MsgDirectionOut = "out"

	MsgStatusSent      = "sent"
	MsgStatusDelivered = "delivered"
	MsgStatusRead      = "read"
)

type msgDoc struct {
	ID         string    `bson:"_id"`
	InstanceID string    `bson:"instance_id"`
	OwnerID    string    `bson:"owner_id"`
	From       string    `bson:"from"`
	Body       string    `bson:"body"`
	Timestamp  time.Time `bson:"timestamp"`
	Direction  string    `bson:"direction"`
	Status     string    `bson:"status"`
}

type MsgDoc struct {
	ID        string    `bson:"_id"        json:"id"`
	From      string    `bson:"from"       json:"from"`
	Body      string    `bson:"body"       json:"body"`
	Timestamp time.Time `bson:"timestamp"  json:"timestamp"`
	Direction string    `bson:"direction"  json:"direction"`
	Status    string    `bson:"status"     json:"status"`
}

func persistMessage(instanceID, ownerID, from, body string, ts time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col := db.Collection(messagesCollection)
	doc := msgDoc{
		ID:         uuid.New().String(),
		InstanceID: instanceID,
		OwnerID:    ownerID,
		From:       from,
		Body:       body,
		Timestamp:  ts,
		Direction:  MsgDirectionIn,
		Status:     MsgStatusDelivered,
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Printf("[%s] persist message: %v", instanceID, err)
	}
}

func PersistSentMessage(instanceID, ownerID, msgWAID, to, body string, ts time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col := db.Collection(messagesCollection)
	doc := msgDoc{
		ID:         msgWAID,
		InstanceID: instanceID,
		OwnerID:    ownerID,
		From:       to,
		Body:       body,
		Timestamp:  ts,
		Direction:  MsgDirectionOut,
		Status:     MsgStatusSent,
	}
	if _, err := col.InsertOne(ctx, doc); err != nil {
		log.Printf("[%s] persist sent message: %v", instanceID, err)
	}
}

func UpdateMessageStatus(instanceID, msgWAID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col := db.Collection(messagesCollection)
	_, err := col.UpdateOne(ctx,
		bson.M{"_id": msgWAID, "instance_id": instanceID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		log.Printf("[%s] update message status: %v", instanceID, err)
	}
}

func ListMessages(ctx context.Context, instanceID, ownerID string, limit int64) ([]MsgDoc, error) {
	col := db.Collection(messagesCollection)
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)
	cursor, err := col.Find(ctx, bson.M{
		"instance_id": instanceID,
		"owner_id":    ownerID,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []MsgDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}
