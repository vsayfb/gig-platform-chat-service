package thread

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Thread struct {
	ID            bson.ObjectID `bson:"_id,omitempty"     json:"id"`
	ParticipantA  string        `bson:"participant_a"     json:"participant_a"`
	ParticipantB  string        `bson:"participant_b"     json:"participant_b"`
	LastMessageAt time.Time     `bson:"last_message_at"   json:"last_message_at"`
	CreatedAt     time.Time     `bson:"created_at"        json:"created_at"`
}

// threadKey returns a canonical (sorted) pair so (A,B) and (B,A) resolve to the same thread.
func threadKey(userA, userB string) (a, b string) {
	if userA < userB {
		return userA, userB
	}
	return userB, userA
}
