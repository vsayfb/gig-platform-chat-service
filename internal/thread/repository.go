package thread

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "threads"

var ErrThreadNotFound = errors.New("thread not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection(collectionName)}
}

// FindOrCreate returns the existing thread for the pair or creates one.
// Thread key is always stored sorted so (A,B) == (B,A).
func (r *Repository) FindOrCreate(ctx context.Context, userA, userB, message string) (*Thread, error) {
	a, b := threadKey(userA, userB)

	filter := bson.M{"participant_a": a, "participant_b": b}

	now := time.Now()

	update := bson.M{
		"$setOnInsert": bson.M{
			"participant_a":   a,
			"participant_b":   b,
			"last_message":    message,
			"last_message_at": now,
			"created_at":      now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var t Thread
	err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Thread, error) {
	var t Thread
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) FindByParticipant(ctx context.Context, userID string) ([]*Thread, error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"participant_a": userID},
			bson.M{"participant_b": userID},
		},
	}

	cursor, err := r.col.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var threads []*Thread

	if err := cursor.All(ctx, &threads); err != nil {
		return nil, err
	}

	return threads, nil
}

// UpdateLastMessageAt bumps the thread timestamp after a new message.
func (r *Repository) UpdateLastMessage(
	ctx context.Context,
	threadID bson.ObjectID,
	message string,
) error {
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": threadID},
		bson.M{
			"$set": bson.M{
				"last_message":    message,
				"last_message_at": time.Now(),
			},
		},
	)

	return err
}

func (r *Repository) IsParticipant(ctx context.Context, threadID bson.ObjectID, userID string) (bool, error) {
	filter := bson.M{
		"_id": threadID,
		"$or": bson.A{
			bson.M{"participant_a": userID},
			bson.M{"participant_b": userID},
		},
	}

	count, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
