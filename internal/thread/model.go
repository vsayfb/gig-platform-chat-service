package thread

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Participant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type ThreadsResponse struct {
	ID            string       `json:"id"`
	Participant   *Participant `json:"participant"`
	LastMessage   string       `json:"last_message"`
	LastMessageAt time.Time    `json:"last_message_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

type Thread struct {
	ID            bson.ObjectID `bson:"_id,omitempty"     json:"id"`
	ParticipantA  string        `bson:"participant_a"     json:"participant_a"`
	ParticipantB  string        `bson:"participant_b"     json:"participant_b"`
	LastMessage   string        `bson:"last_message_"   json:"last_message"`
	LastMessageAt time.Time     `bson:"last_message_at"   json:"last_message_at"`
	CreatedAt     time.Time     `bson:"created_at"        json:"created_at"`
}

type ThreadResponse struct {
	ID            string       `json:"id"`
	Participant   *Participant `json:"participant"`
	LastMessage   string       `json:"last_message"`
	LastMessageAt time.Time    `json:"last_message_at"`
	CreatedAt     time.Time    `json:"created_at"`
}

// threadKey returns a canonical (sorted) pair so (A,B) and (B,A) resolve to the same thread.
func threadKey(userA, userB string) (a, b string) {
	if userA < userB {
		return userA, userB
	}
	return userB, userA
}
