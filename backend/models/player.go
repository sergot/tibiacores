package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Player represents a user account that can have multiple Tibia characters
type Player struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username    string             `json:"username" bson:"username"`
	SessionID   string             `json:"session_id" bson:"session_id"`
	IsAnonymous bool               `json:"is_anonymous" bson:"is_anonymous"`
	Characters  []Character        `json:"characters" bson:"characters"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Character represents a Tibia character belonging to a player
type Character struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	World     string             `json:"world" bson:"world"`
	IsMain    bool               `json:"is_main" bson:"is_main"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
