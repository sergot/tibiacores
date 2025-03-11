package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SoulpitList represents a list of soul cores that players want to complete together
type SoulpitList struct {
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

// ListMember represents a player's character that is part of a soulpit list
type ListMember struct {
	PlayerID       primitive.ObjectID `json:"player_id" bson:"player_id"`
	Username       string             `json:"username" bson:"username"`
	CharacterID    primitive.ObjectID `json:"character_id" bson:"character_id"`
	CharacterName  string             `json:"character_name" bson:"character_name"`
	World          string             `json:"world" bson:"world"`
	SessionID      string             `json:"session_id,omitempty" bson:"session_id,omitempty"`
	IsCreator      bool               `json:"is_creator" bson:"is_creator"`
	JoinedAt       time.Time          `json:"joined_at" bson:"joined_at"`
	SoulCoresAdded int                `json:"soul_cores_added" bson:"soul_cores_added"`
}

// PlayerSoulCore represents a soul core that a specific player has in a list
type PlayerSoulCore struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ListID      primitive.ObjectID `json:"list_id" bson:"list_id"`
	PlayerID    primitive.ObjectID `json:"player_id" bson:"player_id"`
	CharacterID primitive.ObjectID `json:"character_id" bson:"character_id"`
	CreatureID  string             `json:"creature_id" bson:"creature_id"`
	Obtained    bool               `json:"obtained" bson:"obtained"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}
