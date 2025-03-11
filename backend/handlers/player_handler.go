package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/fiendlist/backend/models"
	"github.com/fiendlist/backend/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PlayerHandler handles player-related requests
type PlayerHandler struct {
	DB *mongo.Database
}

// NewPlayerHandler creates a new player handler
func NewPlayerHandler(db *mongo.Database) *PlayerHandler {
	return &PlayerHandler{DB: db}
}

// CreatePlayer creates a new player with their first character
func (h *PlayerHandler) CreatePlayer(c echo.Context) error {
	// Define a request struct to handle the incoming data
	type CreatePlayerRequest struct {
		Username      string `json:"username"`
		CharacterName string `json:"character_name"`
		World         string `json:"world"`
	}

	var req CreatePlayerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
	}
	if req.CharacterName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character name is required"})
	}
	if req.World == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "World is required"})
	}

	// Check if player already exists
	existingPlayer := models.Player{}
	err := h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"username": req.Username},
	).Decode(&existingPlayer)

	// If player exists, add a new character to their account
	if err == nil {
		// Check if character with the same name already exists for this player
		for _, char := range existingPlayer.Characters {
			if char.Name == req.CharacterName {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character already exists"})
			}
		}

		// Create a new character
		newCharacter := models.Character{
			ID:        primitive.NewObjectID(),
			Name:      req.CharacterName,
			World:     req.World,
			IsMain:    len(existingPlayer.Characters) == 0, // First character is main
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generate a new session ID
		sessionID := utils.GenerateSessionID()

		// Update the player with the new character and session ID
		_, err = h.DB.Collection("players").UpdateOne(
			context.Background(),
			bson.M{"_id": existingPlayer.ID},
			bson.M{
				"$set":  bson.M{"session_id": sessionID, "updated_at": time.Now()},
				"$push": bson.M{"characters": newCharacter},
			},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character"})
		}

		// Get the updated player
		updatedPlayer := models.Player{}
		err = h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"_id": existingPlayer.ID},
		).Decode(&updatedPlayer)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated player"})
		}

		return c.JSON(http.StatusOK, updatedPlayer)
	}

	// Create a new player with their first character
	character := models.Character{
		ID:        primitive.NewObjectID(),
		Name:      req.CharacterName,
		World:     req.World,
		IsMain:    true, // First character is main
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	player := models.Player{
		ID:         primitive.NewObjectID(),
		Username:   req.Username,
		SessionID:  utils.GenerateSessionID(),
		Characters: []models.Character{character},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err = h.DB.Collection("players").InsertOne(context.Background(), player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
	}

	return c.JSON(http.StatusCreated, player)
}

// GetPlayerBySessionID gets a player by session ID
func (h *PlayerHandler) GetPlayerBySessionID(c echo.Context) error {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
	}

	var player models.Player
	err := h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"session_id": sessionID},
	).Decode(&player)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusOK, player)
}

// AddCharacter adds a new character to a player's account
func (h *PlayerHandler) AddCharacter(c echo.Context) error {
	playerID := c.Param("playerID")
	if playerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player ID is required"})
	}

	// Convert player ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID"})
	}

	// Define a request struct to handle the incoming data
	type AddCharacterRequest struct {
		CharacterName string `json:"character_name"`
		World         string `json:"world"`
	}

	var req AddCharacterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.CharacterName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character name is required"})
	}
	if req.World == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "World is required"})
	}

	// Get the player
	player := models.Player{}
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&player)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	// Check if character with the same name already exists for this player
	for _, char := range player.Characters {
		if char.Name == req.CharacterName {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character already exists"})
		}
	}

	// Create a new character
	newCharacter := models.Character{
		ID:        primitive.NewObjectID(),
		Name:      req.CharacterName,
		World:     req.World,
		IsMain:    false, // Additional characters are not main by default
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add the character to the player
	_, err = h.DB.Collection("players").UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$push": bson.M{"characters": newCharacter},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character"})
	}

	// Get the updated player
	updatedPlayer := models.Player{}
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&updatedPlayer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated player"})
	}

	return c.JSON(http.StatusOK, updatedPlayer)
}

// SetMainCharacter sets a character as the player's main character
func (h *PlayerHandler) SetMainCharacter(c echo.Context) error {
	playerID := c.Param("playerID")
	characterID := c.Param("characterID")

	if playerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player ID is required"})
	}
	if characterID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character ID is required"})
	}

	// Convert IDs to ObjectIDs
	playerObjectID, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID"})
	}

	characterObjectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid character ID"})
	}

	// First, set all characters to not main
	_, err = h.DB.Collection("players").UpdateOne(
		context.Background(),
		bson.M{"_id": playerObjectID},
		bson.M{
			"$set": bson.M{
				"characters.$[].is_main": false,
				"updated_at":             time.Now(),
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update characters"})
	}

	// Then, set the specified character as main
	_, err = h.DB.Collection("players").UpdateOne(
		context.Background(),
		bson.M{
			"_id":            playerObjectID,
			"characters._id": characterObjectID,
		},
		bson.M{
			"$set": bson.M{
				"characters.$.is_main": true,
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set main character"})
	}

	// Get the updated player
	updatedPlayer := models.Player{}
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": playerObjectID},
	).Decode(&updatedPlayer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated player"})
	}

	return c.JSON(http.StatusOK, updatedPlayer)
}

// ConvertAnonymousToAccount converts an anonymous session to a full account
func (h *PlayerHandler) ConvertAnonymousToAccount(c echo.Context) error {
	// Define a request struct to handle the incoming data
	type ConvertRequest struct {
		SessionID string `json:"session_id"`
		Username  string `json:"username"`
	}

	var req ConvertRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.SessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
	}
	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
	}

	// Check if username is already taken
	var existingPlayer models.Player
	err := h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"username": req.Username},
	).Decode(&existingPlayer)

	if err == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username already taken"})
	} else if err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check username"})
	}

	// Get anonymous session
	var session models.AnonymousSession
	err = h.DB.Collection("anonymous_sessions").FindOne(
		context.Background(),
		bson.M{"session_id": req.SessionID},
	).Decode(&session)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Session not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get session"})
	}

	// Convert anonymous characters to player characters
	playerCharacters := make([]models.Character, 0, len(session.Characters))
	for i, anonChar := range session.Characters {
		playerChar := models.Character{
			ID:        primitive.NewObjectID(),
			Name:      anonChar.Name,
			World:     anonChar.World,
			IsMain:    i == 0, // First character is main
			CreatedAt: anonChar.CreatedAt,
			UpdatedAt: time.Now(),
		}
		playerCharacters = append(playerCharacters, playerChar)
	}

	// Create new player
	player := models.Player{
		ID:         primitive.NewObjectID(),
		Username:   req.Username,
		SessionID:  utils.GenerateSessionID(),
		Characters: playerCharacters,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  time.Now(),
	}

	// Insert player
	_, err = h.DB.Collection("players").InsertOne(context.Background(), player)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
	}

	// Update lists to reference the new player ID
	for _, listID := range session.Lists {
		// Get the list
		var list models.SoulpitList
		err = h.DB.Collection("soulpit_lists").FindOne(
			context.Background(),
			bson.M{"_id": listID},
		).Decode(&list)

		if err != nil {
			continue // Skip if list not found
		}

		// Update creator ID if this session is the creator
		updateFields := bson.M{"updated_at": time.Now()}
		if list.CreatorID == session.ID {
			updateFields["creator_id"] = player.ID
		}

		// Update members
		for i, member := range list.Members {
			if member.PlayerID == session.ID {
				// Find the corresponding player character
				for _, playerChar := range playerCharacters {
					if playerChar.Name == member.CharacterName {
						list.Members[i].PlayerID = player.ID
						list.Members[i].CharacterID = playerChar.ID
						break
					}
				}
			}
		}

		// Update the list
		_, err = h.DB.Collection("soulpit_lists").UpdateOne(
			context.Background(),
			bson.M{"_id": listID},
			bson.M{
				"$set": bson.M{
					"creator_id": updateFields["creator_id"],
					"members":    list.Members,
					"updated_at": updateFields["updated_at"],
				},
			},
		)
	}

	// Delete anonymous session
	_, err = h.DB.Collection("anonymous_sessions").DeleteOne(
		context.Background(),
		bson.M{"_id": session.ID},
	)

	return c.JSON(http.StatusOK, player)
}

// GetPlayerBySession gets a player by session ID
func (h *PlayerHandler) GetPlayerBySession(c echo.Context) error {
	sessionID := c.Param("sessionID")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Session ID is required"})
	}

	// Find player by session ID
	var player models.Player
	err := h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"session_id": sessionID},
	).Decode(&player)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	return c.JSON(http.StatusOK, player)
}

// GetCharacters gets all characters for a player
func (h *PlayerHandler) GetCharacters(c echo.Context) error {
	playerID := c.Param("playerID")
	if playerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player ID is required"})
	}

	// Convert player ID to ObjectID
	playerObjectID, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID format"})
	}

	// Find player by ID
	var player models.Player
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": playerObjectID},
	).Decode(&player)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	// Return just the characters array
	return c.JSON(http.StatusOK, player.Characters)
}

// UpdateUsername updates a player's username
func (h *PlayerHandler) UpdateUsername(c echo.Context) error {
	playerID := c.Param("playerID")
	if playerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player ID is required"})
	}

	// Convert playerID to ObjectID
	playerObjectID, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID format"})
	}

	// Define a request struct to handle the incoming data
	type UpdateUsernameRequest struct {
		Username string `json:"username"`
	}

	var req UpdateUsernameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate required fields
	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username is required"})
	}

	// Check if username is already taken by another player
	var existingPlayer models.Player
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{
			"username": req.Username,
			"_id":      bson.M{"$ne": playerObjectID}, // Not the current player
		},
	).Decode(&existingPlayer)

	if err == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username already taken"})
	} else if err != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check username"})
	}

	// Update the player's username
	update := bson.M{
		"$set": bson.M{
			"username":   req.Username,
			"updated_at": time.Now(),
		},
	}

	_, err = h.DB.Collection("players").UpdateOne(
		context.Background(),
		bson.M{"_id": playerObjectID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update username"})
	}

	// Get the updated player
	updatedPlayer := models.Player{}
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": playerObjectID},
	).Decode(&updatedPlayer)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated player"})
	}

	return c.JSON(http.StatusOK, updatedPlayer)
}
