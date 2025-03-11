package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnonymousSession represents a temporary session for users without accounts
type AnonymousSession struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	SessionID  string               `json:"session_id" bson:"session_id"`
	Characters []AnonymousCharacter `json:"characters" bson:"characters"`
	Lists      []primitive.ObjectID `json:"lists" bson:"lists"`
	CreatedAt  time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at" bson:"updated_at"`
}

// AnonymousCharacter represents a character created by an anonymous user
type AnonymousCharacter struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	World     string             `json:"world" bson:"world"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
