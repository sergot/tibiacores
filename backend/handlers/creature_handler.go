package handlers

import (
	"context"
	"net/http"

	"github.com/fiendlist/backend/models"
	"github.com/fiendlist/backend/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreatureHandler handles creature-related requests
type CreatureHandler struct {
	DB *mongo.Database
}

// NewCreatureHandler creates a new creature handler
func NewCreatureHandler(db *mongo.Database) *CreatureHandler {
	return &CreatureHandler{DB: db}
}

// ImportCreatures imports creatures from the JSON file
func (h *CreatureHandler) ImportCreatures(c echo.Context) error {
	// Load creatures from JSON file
	creaturesData, err := utils.LoadCreaturesFromJSON("creatures.json")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load creatures"})
	}

	// Extract creatures array
	creaturesArray, ok := creaturesData["creatures"].([]interface{})
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid creatures data format"})
	}

	// Insert creatures into database
	for _, creatureData := range creaturesArray {
		creatureMap, ok := creatureData.(map[string]interface{})
		if !ok {
			continue
		}

		creature := models.Creature{
			Endpoint:   creatureMap["endpoint"].(string),
			PluralName: creatureMap["plural_name"].(string),
			Name:       creatureMap["name"].(string),
		}

		// Use upsert to avoid duplicates
		_, err := h.DB.Collection("creatures").UpdateOne(
			context.Background(),
			bson.M{"endpoint": creature.Endpoint},
			bson.M{"$set": creature},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to import creatures"})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Creatures imported successfully"})
}

// GetAllCreatures gets all creatures
func (h *CreatureHandler) GetAllCreatures(c echo.Context) error {
	var creatures []models.Creature

	cursor, err := h.DB.Collection("creatures").Find(context.Background(), bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get creatures"})
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &creatures); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode creatures"})
	}

	return c.JSON(http.StatusOK, creatures)
}

// GetCreatureByEndpoint gets a creature by endpoint
func (h *CreatureHandler) GetCreatureByEndpoint(c echo.Context) error {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Endpoint is required"})
	}

	var creature models.Creature
	err := h.DB.Collection("creatures").FindOne(
		context.Background(),
		bson.M{"endpoint": endpoint},
	).Decode(&creature)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Creature not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get creature"})
	}

	return c.JSON(http.StatusOK, creature)
}
