package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// List represents a list of soul cores that players want to complete together
type List struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	World       string             `json:"world" bson:"world"`
	CreatorID   primitive.ObjectID `json:"creator_id" bson:"creator_id"`
	Members     []ListMember       `json:"members" bson:"members"`
	SoulCores   []SoulCore         `json:"soul_cores" bson:"soul_cores"`
	ShareCode   string             `json:"share_code" bson:"share_code"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
