package message

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Message struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ThreadID bson.ObjectID `bson:"thread_id"     json:"thread_id"`
	SenderID string        `bson:"sender_id"     json:"sender_id"`
	Content  string        `bson:"content"       json:"content"`
	IsRead   bool          `bson:"is_read"       json:"is_read"`
	SentAt   time.Time     `bson:"sent_at"       json:"sent_at"`
}
