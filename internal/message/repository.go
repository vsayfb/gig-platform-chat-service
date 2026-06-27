package message

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "messages"

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection(collectionName)}
}

func (r *Repository) Insert(ctx context.Context, msg *Message) error {
	msg.ID = bson.NewObjectID()
	msg.SentAt = time.Now()
	msg.IsRead = false

	_, err := r.col.InsertOne(ctx, msg)
	return err
}

func (r *Repository) FindByThread(ctx context.Context, threadID bson.ObjectID, cursor time.Time, limit int) ([]*Message, error) {
	filter := bson.M{"thread_id": threadID}

	if !cursor.IsZero() {
		filter["sent_at"] = bson.M{"$lt": cursor}
	}

	opts := options.Find().
		SetSort(bson.M{"sent_at": -1}).
		SetLimit(int64(limit))

	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var msgs []*Message
	if err := cur.All(ctx, &msgs); err != nil {
		return nil, err
	}

	return msgs, nil
}
