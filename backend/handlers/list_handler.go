package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/fiendlist/backend/models"
	"github.com/fiendlist/backend/utils"
)

// ListHandler handles list-related requests
type ListHandler struct {
	DB *mongo.Database
}

// NewListHandler creates a new list handler
func NewListHandler(db *mongo.Database) *ListHandler {
	return &ListHandler{DB: db}
}

// CreateList creates a new list
func (h *ListHandler) CreateList(c echo.Context) error {
	var req struct {
		Name          string `json:"name" validate:"required"`
		Description   string `json:"description"`
		PlayerID      string `json:"player_id"`
		CharacterID   string `json:"character_id"`
		CharacterName string `json:"character_name"`
		World         string `json:"world"`
		SessionID     string `json:"session_id"`
	}

	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	log.Printf("Received request: %+v", req)

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	var playerID primitive.ObjectID
	var characterID primitive.ObjectID
	var characterName string
	var world string

	// Scenario 1: User with player_id and character_name (creating a list with a new character)
	if req.PlayerID != "" && req.CharacterName != "" && req.World != "" && req.CharacterID == "" {
		log.Printf("Scenario 1A: Creating new character for existing player ID: %s, Character: %s, World: %s", req.PlayerID, req.CharacterName, req.World)

		var err error
		playerID, err = primitive.ObjectIDFromHex(req.PlayerID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID"})
		}

		// Get player to verify existence
		var player models.Player
		err = h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"_id": playerID},
		).Decode(&player)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
		}

		// Create a new character for the player
		character := models.Character{
			ID:        primitive.NewObjectID(),
			Name:      req.CharacterName,
			World:     req.World,
			IsMain:    false, // Not setting as main since player already has characters
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Add character to player
		player.Characters = append(player.Characters, character)
		player.UpdatedAt = time.Now()

		// Update player in database
		_, err = h.DB.Collection("players").ReplaceOne(
			context.Background(),
			bson.M{"_id": playerID},
			player,
		)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character to player"})
		}

		characterID = character.ID
		characterName = req.CharacterName
		world = req.World

		// Scenario 1: User with player_id (returning user)
	} else if req.PlayerID != "" && req.CharacterID == "" {
		log.Printf("Scenario 1B: Using player ID: %s", req.PlayerID)

		var err error
		playerID, err = primitive.ObjectIDFromHex(req.PlayerID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID"})
		}

		// Get player to verify existence
		var player models.Player
		err = h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"_id": playerID},
		).Decode(&player)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
		}

		// Use the main character or first character
		if len(player.Characters) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player has no characters"})
		}

		// Try to find the main character
		mainCharFound := false
		for _, char := range player.Characters {
			if char.IsMain {
				characterID = char.ID
				characterName = char.Name
				world = char.World
				mainCharFound = true
				break
			}
		}

		// If no main character, use the first character
		if !mainCharFound {
			characterID = player.Characters[0].ID
			characterName = player.Characters[0].Name
			world = player.Characters[0].World
		}

		// Scenario 2: First-time user with session_id
	} else if req.SessionID != "" && req.CharacterID == "" {
		log.Printf("Scenario 2: First-time user with session ID: %s", req.SessionID)

		if req.CharacterName == "" || req.World == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character name and world are required for first-time users"})
		}

		// Check if a player with this session ID already exists
		var existingPlayer models.Player
		err := h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"session_id": req.SessionID},
		).Decode(&existingPlayer)

		if err == mongo.ErrNoDocuments {
			// Create new anonymous player
			player := models.Player{
				ID:          primitive.NewObjectID(),
				Username:    req.CharacterName,
				SessionID:   req.SessionID,
				IsAnonymous: true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Create character for the player
			character := models.Character{
				ID:        primitive.NewObjectID(),
				Name:      req.CharacterName,
				World:     req.World,
				IsMain:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Add character to player
			player.Characters = []models.Character{character}

			// Insert player into database
			_, err := h.DB.Collection("players").InsertOne(context.Background(), player)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
			}

			playerID = player.ID
			characterID = character.ID
			characterName = req.CharacterName
			world = req.World
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check for existing player"})
		} else {
			// This shouldn't happen in this scenario (player already exists with this session ID)
			// But handle it gracefully by using the existing player
			log.Printf("Warning: Player already exists with session ID %s", req.SessionID)

			playerID = existingPlayer.ID

			// Use the main character or first character
			if len(existingPlayer.Characters) == 0 {
				// Create a new character if none exists
				character := models.Character{
					ID:        primitive.NewObjectID(),
					Name:      req.CharacterName,
					World:     req.World,
					IsMain:    true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				// Add character to player
				existingPlayer.Characters = append(existingPlayer.Characters, character)
				existingPlayer.UpdatedAt = time.Now()

				_, err := h.DB.Collection("players").ReplaceOne(
					context.Background(),
					bson.M{"_id": existingPlayer.ID},
					existingPlayer,
				)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character to player"})
				}

				characterID = character.ID
				characterName = req.CharacterName
				world = req.World
			} else {
				// Try to find the main character
				mainCharFound := false
				for _, char := range existingPlayer.Characters {
					if char.IsMain {
						characterID = char.ID
						characterName = char.Name
						world = char.World
						mainCharFound = true
						break
					}
				}

				// If no main character, use the first character
				if !mainCharFound {
					characterID = existingPlayer.Characters[0].ID
					characterName = existingPlayer.Characters[0].Name
					world = existingPlayer.Characters[0].World
				}
			}
		}

		// Scenario 3: User with character_id (creating another list with specific character)
	} else if req.CharacterID != "" {
		log.Printf("Scenario 3: Using character ID: %s", req.CharacterID)

		characterObjectID, err := primitive.ObjectIDFromHex(req.CharacterID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid character ID"})
		}

		// Find the player that owns this character
		var player models.Player
		err = h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"characters._id": characterObjectID},
		).Decode(&player)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found for this character"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
		}

		// Find the character in the player's characters
		characterFound := false
		var character models.Character
		for _, char := range player.Characters {
			if char.ID == characterObjectID {
				character = char
				characterFound = true
				break
			}
		}

		if !characterFound {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character not found for this player"})
		}

		playerID = player.ID
		characterID = characterObjectID
		characterName = character.Name
		world = character.World
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either player_id, character_id, or session_id is required"})
	}

	// Get the player's username
	var username string
	if !playerID.IsZero() {
		var playerInfo models.Player
		err := h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"_id": playerID},
		).Decode(&playerInfo)

		if err == nil {
			username = playerInfo.Username
		} else {
			username = characterName // Fallback to character name if player not found
		}
	} else {
		username = characterName // For anonymous users, use character name
	}

	// Create the list member
	member := models.ListMember{
		PlayerID:       playerID,
		Username:       username,
		CharacterID:    characterID,
		CharacterName:  characterName,
		World:          world,
		SessionID:      req.SessionID,
		IsCreator:      true,
		JoinedAt:       time.Now(),
		SoulCoresAdded: 0,
	}

	// Generate a share code
	shareCode := utils.GenerateShareCode()

	// Create the list
	list := models.List{
		ID:          primitive.NewObjectID(),
		Name:        req.Name,
		Description: req.Description,
		World:       world,
		CreatorID:   playerID,
		Members:     []models.ListMember{member},
		SoulCores:   []models.SoulCore{},
		ShareCode:   shareCode,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := h.DB.Collection("lists").InsertOne(context.Background(), list)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create list"})
	}

	return c.JSON(http.StatusCreated, list)
}

// GetLists gets all lists for a user by player ID
func (h *ListHandler) GetLists(c echo.Context) error {
	var lists []models.List

	// Get player ID from query parameter
	playerID := c.QueryParam("player_id")

	// If no player ID provided, try to get it from session
	if playerID == "" {
		// Try to get player ID from session as fallback
		playerIDInterface := c.Get("playerID")
		if playerIDInterface != nil {
			if id, ok := playerIDInterface.(string); ok && id != "" {
				playerID = id
				log.Printf("Using player ID from session: %s", playerID)
			}
		}

		// If still no player ID, return empty list
		if playerID == "" {
			log.Printf("No player ID provided, returning empty list")
			return c.JSON(http.StatusOK, lists)
		}
	}

	// Get lists for the player
	log.Printf("Getting lists for player ID: %s", playerID)

	playerObjectID, err := primitive.ObjectIDFromHex(playerID)
	if err != nil {
		log.Printf("Error converting player ID to ObjectID: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID format"})
	}

	// Find all lists where the player is a member
	cursor, err := h.DB.Collection("lists").Find(
		context.Background(),
		bson.M{"members.player_id": playerObjectID},
	)
	if err != nil {
		log.Printf("Error finding lists for player ID %s: %v", playerID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get lists"})
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &lists); err != nil {
		log.Printf("Error decoding lists for player ID %s: %v", playerID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode lists"})
	}

	log.Printf("Found %d lists for player ID %s", len(lists), playerID)

	return c.JSON(http.StatusOK, lists)
}

// JoinList adds a user to a list (both authenticated and anonymous)
func (h *ListHandler) JoinList(c echo.Context) error {
	var req struct {
		Token       string `json:"token" validate:"required"`
		PlayerID    string `json:"player_id"`
		CharacterID string `json:"character_id"`
		// Anonymous user fields
		CharacterName string `json:"character_name"`
		World         string `json:"world"`
		SessionID     string `json:"session_id"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Token is required"})
	}

	// Find the list by share code
	var list models.List
	err := h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"share_code": req.Token},
	).Decode(&list)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get list"})
	}

	// Check if the list already has 5 members
	if len(list.Members) >= 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "This list has reached the maximum number of members (5)"})
	}

	var member models.ListMember
	var playerID primitive.ObjectID
	var characterID primitive.ObjectID
	var characterName string
	var world string
	var sessionID string

	// Handle authenticated user with player ID
	if req.PlayerID != "" {
		playerObjectID, err := primitive.ObjectIDFromHex(req.PlayerID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid player ID"})
		}

		// Get player to verify existence
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

		// If character ID is provided, verify it belongs to this player
		if req.CharacterID != "" {
			characterObjectID, err := primitive.ObjectIDFromHex(req.CharacterID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid character ID"})
			}

			characterFound := false
			var character models.Character
			for _, char := range player.Characters {
				if char.ID == characterObjectID {
					character = char
					characterFound = true
					break
				}
			}
			if !characterFound {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character not found for this player"})
			}

			// Check if the player is already a member of the list with this character
			for _, existingMember := range list.Members {
				if existingMember.PlayerID == playerObjectID && existingMember.CharacterID == characterObjectID {
					return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character is already a member of this list"})
				}
			}

			playerID = playerObjectID
			characterID = characterObjectID
			characterName = character.Name
			world = character.World
		} else if req.CharacterName != "" && req.World != "" {
			// Handle existing player with new character
			log.Printf("Handling existing player with new character: Player ID: %s, Character: %s, World: %s",
				req.PlayerID, req.CharacterName, req.World)

			// Check if the player already has this character
			characterExists := false
			var existingCharacter models.Character

			for _, char := range player.Characters {
				if strings.EqualFold(char.Name, req.CharacterName) && strings.EqualFold(char.World, req.World) {
					characterExists = true
					existingCharacter = char
					break
				}
			}

			if characterExists {
				// Use existing character
				characterID = existingCharacter.ID
				characterName = existingCharacter.Name
				world = existingCharacter.World

				// Check if the player is already a member of the list with this character
				for _, existingMember := range list.Members {
					if existingMember.PlayerID == playerObjectID && existingMember.CharacterID == characterID {
						return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character is already a member of this list"})
					}
				}
			} else {
				// Create a new character for the player
				newCharacter := models.Character{
					ID:        primitive.NewObjectID(),
					Name:      req.CharacterName,
					World:     req.World,
					IsMain:    len(player.Characters) == 0, // Only set as main if this is the first character
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				// Add character to player
				player.Characters = append(player.Characters, newCharacter)
				player.UpdatedAt = time.Now()

				// Update player in database
				_, err = h.DB.Collection("players").ReplaceOne(
					context.Background(),
					bson.M{"_id": playerObjectID},
					player,
				)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character to player"})
				}

				characterID = newCharacter.ID
				characterName = newCharacter.Name
				world = newCharacter.World
			}

			playerID = playerObjectID
		} else {
			// If no character ID or character name is provided, use the main character
			if len(player.Characters) == 0 {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player has no characters"})
			}

			var character models.Character
			mainCharFound := false
			for _, char := range player.Characters {
				if char.IsMain {
					character = char
					mainCharFound = true
					break
				}
			}

			if !mainCharFound {
				character = player.Characters[0]
			}

			// Check if the player is already a member of the list with this character
			for _, existingMember := range list.Members {
				if existingMember.PlayerID == playerObjectID && existingMember.CharacterID == character.ID {
					return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character is already a member of this list"})
				}
			}

			playerID = playerObjectID
			characterID = character.ID
			characterName = character.Name
			world = character.World
		}
	} else if req.CharacterID != "" {
		// User with character ID but no player ID
		characterObjectID, err := primitive.ObjectIDFromHex(req.CharacterID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid character ID"})
		}

		// Find the player that owns this character
		var player models.Player
		err = h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"characters._id": characterObjectID},
		).Decode(&player)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found for this character"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
		}

		// Find the character in the player's characters
		characterFound := false
		var character models.Character
		for _, char := range player.Characters {
			if char.ID == characterObjectID {
				character = char
				characterFound = true
				break
			}
		}

		if !characterFound {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character not found for this player"})
		}

		// Check if the player is already a member of the list with this character
		for _, existingMember := range list.Members {
			if existingMember.PlayerID == player.ID && existingMember.CharacterID == characterObjectID {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character is already a member of this list"})
			}
		}

		playerID = player.ID
		characterID = characterObjectID
		characterName = character.Name
		world = character.World
	} else if req.SessionID != "" {
		// Anonymous user with session ID
		if req.CharacterName == "" || req.World == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character details are required for anonymous users"})
		}

		// Check if a player with this session ID already exists
		var existingPlayer models.Player
		err := h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"session_id": req.SessionID},
		).Decode(&existingPlayer)

		if err == mongo.ErrNoDocuments {
			// Create new player for anonymous user
			player := models.Player{
				ID:          primitive.NewObjectID(),
				Username:    req.CharacterName,
				SessionID:   req.SessionID,
				IsAnonymous: true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Create character for anonymous user
			character := models.Character{
				ID:        primitive.NewObjectID(),
				Name:      req.CharacterName,
				World:     req.World,
				IsMain:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Add character to player
			player.Characters = []models.Character{character}

			// Insert player into database
			_, err := h.DB.Collection("players").InsertOne(context.Background(), player)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create player"})
			}

			playerID = player.ID
			characterID = character.ID
			characterName = req.CharacterName
			world = req.World
			sessionID = req.SessionID
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check for existing player"})
		} else {
			// Player exists, check if character exists
			characterExists := false
			for _, char := range existingPlayer.Characters {
				if char.Name == req.CharacterName && char.World == req.World {
					characterID = char.ID
					characterExists = true
					break
				}
			}

			if !characterExists {
				// Create new character for existing player
				newCharacter := models.Character{
					ID:        primitive.NewObjectID(),
					Name:      req.CharacterName,
					World:     req.World,
					IsMain:    len(existingPlayer.Characters) == 0,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}

				characterID = newCharacter.ID
				characterName = req.CharacterName
				world = req.World

				// Add character to player
				existingPlayer.Characters = append(existingPlayer.Characters, newCharacter)
				existingPlayer.UpdatedAt = time.Now()

				_, err := h.DB.Collection("players").ReplaceOne(
					context.Background(),
					bson.M{"_id": existingPlayer.ID},
					existingPlayer,
				)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add character to player"})
				}
			} else {
				characterName = req.CharacterName
				world = req.World
			}

			playerID = existingPlayer.ID
			sessionID = req.SessionID
		}

		// Check if the player is already a member of the list with this character
		for _, existingMember := range list.Members {
			if existingMember.PlayerID == playerID && existingMember.CharacterID == characterID {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character is already a member of this list"})
			}
		}
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Either player_id, character_id, or session_id is required"})
	}

	// Get the player's username
	var username string
	if !playerID.IsZero() {
		var playerInfo models.Player
		err := h.DB.Collection("players").FindOne(
			context.Background(),
			bson.M{"_id": playerID},
		).Decode(&playerInfo)

		if err == nil {
			username = playerInfo.Username
		} else {
			username = characterName // Fallback to character name if player not found
		}
	} else {
		username = characterName // For anonymous users, use character name
	}

	// Create the list member
	member = models.ListMember{
		PlayerID:       playerID,
		Username:       username,
		CharacterID:    characterID,
		CharacterName:  characterName,
		World:          world,
		SessionID:      sessionID,
		IsCreator:      false,
		JoinedAt:       time.Now(),
		SoulCoresAdded: 0,
	}

	// Add member to list
	_, err = h.DB.Collection("lists").UpdateOne(
		context.Background(),
		bson.M{"_id": list.ID},
		bson.M{
			"$push": bson.M{"members": member},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add member to list"})
	}

	// Get the updated list
	err = h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"_id": list.ID},
	).Decode(&list)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated list"})
	}

	return c.JSON(http.StatusOK, list)
}

// GetListByID gets a list by ID
func (h *ListHandler) GetListByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "List ID is required"})
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}

	var list models.List
	err = h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&list)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get list"})
	}

	return c.JSON(http.StatusOK, list)
}

// GetListByShareCode gets a list by share code
func (h *ListHandler) GetListByShareCode(c echo.Context) error {
	shareCode := c.Param("shareCode")
	if shareCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Share code is required"})
	}

	var list models.List
	err := h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"share_code": shareCode},
	).Decode(&list)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get list"})
	}

	return c.JSON(http.StatusOK, list)
}

// AddSoulCoreToList adds a soul core to a list
func (h *ListHandler) AddSoulCoreToList(c echo.Context) error {
	listID := c.Param("listID")
	if listID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "List ID is required"})
	}

	// Convert list ID to ObjectID
	listObjectID, err := primitive.ObjectIDFromHex(listID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}

	// Define a request struct to handle the incoming data
	type AddSoulCoreRequest struct {
		CreatureID string             `json:"creature_id"`
		PlayerID   primitive.ObjectID `json:"player_id"`
	}

	var requestBody AddSoulCoreRequest
	if err := c.Bind(&requestBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate required fields
	if requestBody.CreatureID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Creature ID is required"})
	}
	if requestBody.PlayerID.IsZero() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player ID is required"})
	}

	// Get the list
	var list models.List
	err = h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"_id": listObjectID},
	).Decode(&list)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get list"})
	}

	// Check if the player is a member of the list
	playerIsMember := false
	var memberIndex int
	for i, member := range list.Members {
		if member.PlayerID == requestBody.PlayerID {
			playerIsMember = true
			memberIndex = i
			break
		}
	}
	if !playerIsMember {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Player is not a member of this list"})
	}

	// Check if the soul core already exists in the list
	for _, soulCore := range list.SoulCores {
		if soulCore.CreatureID == requestBody.CreatureID {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Soul core already exists in the list"})
		}
	}

	// Get the creature
	var creature models.Creature
	err = h.DB.Collection("creatures").FindOne(
		context.Background(),
		bson.M{"endpoint": requestBody.CreatureID},
	).Decode(&creature)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Creature not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get creature"})
	}

	// Get the player to get the character name
	player := models.Player{}
	err = h.DB.Collection("players").FindOne(
		context.Background(),
		bson.M{"_id": requestBody.PlayerID},
	).Decode(&player)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Player not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get player"})
	}

	// Find the character name and ID
	var characterName string
	var characterID primitive.ObjectID

	// Find the member in the list to get the character ID
	var memberCharacterID primitive.ObjectID
	var memberCharacterName string

	for _, member := range list.Members {
		if member.PlayerID == requestBody.PlayerID {
			memberCharacterID = member.CharacterID
			memberCharacterName = member.CharacterName
			break
		}
	}

	// If we found a member, use their character ID and name
	if !memberCharacterID.IsZero() {
		characterID = memberCharacterID
		characterName = memberCharacterName
	} else if len(player.Characters) > 0 {
		// Fallback to the first character if no member found
		characterID = player.Characters[0].ID
		characterName = player.Characters[0].Name
	} else {
		// Last resort fallback
		characterName = player.Username
		if characterName == "" {
			characterName = "Unknown Player"
		}
		// Generate a placeholder character ID
		characterID = primitive.NewObjectID()
	}

	// Create the soul core
	soulCore := models.SoulCore{
		ID:             primitive.NewObjectID(),
		CreatureID:     requestBody.CreatureID,
		Creature:       creature,
		Obtained:       true, // Soul cores are obtained when added
		ObtainedBy:     characterID,
		ObtainedByName: characterName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Add the soul core to the list
	_, err = h.DB.Collection("lists").UpdateOne(
		context.Background(),
		bson.M{"_id": listObjectID},
		bson.M{
			"$push": bson.M{"soul_cores": soulCore},
			"$set":  bson.M{"updated_at": time.Now()},
			"$inc":  bson.M{fmt.Sprintf("members.%d.soul_cores_added", memberIndex): 1},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add soul core to list"})
	}

	return c.JSON(http.StatusCreated, soulCore)
}

// UpdateSoulCoreInList updates a soul core in a list
func (h *ListHandler) UpdateSoulCoreInList(c echo.Context) error {
	listID := c.Param("listID")
	soulCoreID := c.Param("soulCoreID")
	if listID == "" || soulCoreID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "List ID and Soul Core ID are required"})
	}

	var requestBody struct {
		Obtained bool `json:"obtained"`
		Unlocked bool `json:"unlocked"`
	}

	if err := c.Bind(&requestBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	listObjectID, err := primitive.ObjectIDFromHex(listID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid list ID"})
	}

	soulCoreIDObjectID, err := primitive.ObjectIDFromHex(soulCoreID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid soul core ID"})
	}

	var list models.List
	err = h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"_id": listObjectID},
	).Decode(&list)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "List not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get list"})
	}

	// Find the soul core in the list
	var soulCoreIndex = -1
	for i, sc := range list.SoulCores {
		if sc.ID == soulCoreIDObjectID {
			soulCoreIndex = i
			break
		}
	}
	if soulCoreIndex == -1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Soul core not found in the list"})
	}

	// Update the soul core fields
	updateFields := bson.M{}

	// Only update fields that are provided in the request
	if requestBody.Obtained {
		updateFields[fmt.Sprintf("soul_cores.%d.obtained", soulCoreIndex)] = true
	}

	if requestBody.Unlocked {
		updateFields[fmt.Sprintf("soul_cores.%d.unlocked", soulCoreIndex)] = true
	}

	updateFields[fmt.Sprintf("soul_cores.%d.updated_at", soulCoreIndex)] = time.Now()

	_, err = h.DB.Collection("lists").UpdateOne(
		context.Background(),
		bson.M{"_id": listObjectID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update soul core in the list"})
	}

	// Get the updated list
	var updatedList models.List
	err = h.DB.Collection("lists").FindOne(
		context.Background(),
		bson.M{"_id": listObjectID},
	).Decode(&updatedList)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get updated list"})
	}

	// Find the updated soul core
	var updatedSoulCore models.SoulCore
	for _, sc := range updatedList.SoulCores {
		if sc.ID == soulCoreIDObjectID {
			updatedSoulCore = sc
			break
		}
	}

	return c.JSON(http.StatusOK, updatedSoulCore)
}

// GetListsByCharacterID retrieves all lists where a specific character is a member
func (h *ListHandler) GetListsByCharacterID(c echo.Context) error {
	characterID := c.Param("characterID")
	if characterID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Character ID is required"})
	}

	// Convert characterID to ObjectID
	characterObjectID, err := primitive.ObjectIDFromHex(characterID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid character ID format"})
	}

	// Find the player that has this character
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"characters._id": characterObjectID}}},
		{{"$project", bson.M{
			"_id":      1,
			"username": 1,
			"characters": bson.M{"$filter": bson.M{
				"input": "$characters",
				"as":    "character",
				"cond":  bson.M{"$eq": []interface{}{"$$character._id", characterObjectID}},
			}},
		}}},
	}

	cursor, err := h.DB.Collection("players").Aggregate(context.Background(), pipeline)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find player with character"})
	}
	defer cursor.Close(context.Background())

	var players []models.Player
	if err = cursor.All(context.Background(), &players); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode player data"})
	}

	if len(players) == 0 || len(players[0].Characters) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Character not found"})
	}

	// Get the character from the player
	character := players[0].Characters[0]

	// Find all lists where this character is a member
	cursor, err = h.DB.Collection("lists").Find(
		context.Background(),
		bson.M{
			"members": bson.M{
				"$elemMatch": bson.M{
					"character_name": character.Name,
					"world":          character.World,
				},
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get lists"})
	}
	defer cursor.Close(context.Background())

	var lists []models.List
	if err = cursor.All(context.Background(), &lists); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode lists"})
	}

	// Process lists to include only relevant information
	type SoulCoreInfo struct {
		ID           primitive.ObjectID `json:"id"`
		CreatureName string             `json:"creature_name"`
		Unlocked     bool               `json:"unlocked"`
		ListID       primitive.ObjectID `json:"list_id"`
		ListName     string             `json:"list_name"`
	}

	var unlockedSoulCores []SoulCoreInfo
	for _, list := range lists {
		for _, soulCore := range list.SoulCores {
			if soulCore.Unlocked {
				// Find the creature name
				var creatureName string
				if soulCore.Creature.Name != "" {
					creatureName = soulCore.Creature.Name
				} else {
					creatureName = "Unknown"
				}

				unlockedSoulCores = append(unlockedSoulCores, SoulCoreInfo{
					ID:           soulCore.ID,
					CreatureName: creatureName,
					Unlocked:     true,
					ListID:       list.ID,
					ListName:     list.Name,
				})
			}
		}
	}

	// Remove duplicates based on creature name
	uniqueSoulCores := []SoulCoreInfo{}
	creatureNames := make(map[string]bool)

	for _, core := range unlockedSoulCores {
		if !creatureNames[core.CreatureName] {
			creatureNames[core.CreatureName] = true
			uniqueSoulCores = append(uniqueSoulCores, core)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"character":  character,
		"soul_cores": uniqueSoulCores,
	})
}
