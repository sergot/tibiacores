package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SoulCore represents a soul core that a player has
type SoulCore struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatureID     string             `json:"creature_id" bson:"creature_id"`
	Creature       Creature           `json:"creature" bson:"creature"`
	Obtained       bool               `json:"obtained" bson:"obtained"`
	Unlocked       bool               `json:"unlocked" bson:"unlocked"`
	ObtainedBy     primitive.ObjectID `json:"obtained_by,omitempty" bson:"obtained_by,omitempty"`
	ObtainedByName string             `json:"obtained_by_name,omitempty" bson:"obtained_by_name,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
}
