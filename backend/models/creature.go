package models

// Creature represents a creature from Tibia
type Creature struct {
	Endpoint   string `json:"endpoint" bson:"endpoint"`
	PluralName string `json:"plural_name" bson:"plural_name"`
	Name       string `json:"name" bson:"name"`
}
