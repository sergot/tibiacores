package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/services"
)

type ClaimsHandler struct {
	connPool  *pgxpool.Pool
	tibiaData *services.TibiaDataService
}

type StartClaimResponse struct {
	ClaimID          string `json:"claim_id"`
	VerificationCode string `json:"verification_code"`
	Status           string `json:"status"`
	ClaimerID        string `json:"claimer_id,omitempty"` // ID of the claiming user
}

func NewClaimsHandler(connPool *pgxpool.Pool) *ClaimsHandler {
	return &ClaimsHandler{
		connPool:  connPool,
		tibiaData: services.NewTibiaDataService(),
	}
}

// StartClaim initiates a character claim process
func (h *ClaimsHandler) StartClaim(c echo.Context) error {
	var req struct {
		CharacterName string `json:"character_name"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// First check if character exists in TibiaData API
	tibiaChar, err := h.tibiaData.GetCharacter(req.CharacterName)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "character not found in Tibia")
	}

	// Check if character exists in our database
	character, err := queries.GetCharacterByName(ctx, tibiaChar.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "character is not registered in any list yet")
	}

	// Get or create user
	var userID uuid.UUID
	var token string
	// Check if user is authenticated
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
		}
	} else {
		// Create new anonymous user account
		newUser, err := queries.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create user")
		}
		userID = newUser.ID

		// Generate token
		token, err = auth.GenerateToken(userID.String(), false)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
		}

		c.Response().Header().Set("X-Auth-Token", token)
	}

	// Check if there's already an active claim
	existingClaim, err := queries.GetCharacterClaim(ctx, db.GetCharacterClaimParams{
		CharacterID: character.ID,
		ClaimerID:   userID,
	})
	if err == nil && existingClaim.Status == "pending" {
		return c.JSON(http.StatusOK, StartClaimResponse{
			ClaimID:          existingClaim.ID.String(),
			VerificationCode: existingClaim.VerificationCode,
			Status:           existingClaim.Status,
		})
	}

	// Generate random verification code
	verificationBytes := make([]byte, 16)
	if _, err := rand.Read(verificationBytes); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate verification code")
	}
	verificationCode := "TIBIACORES-" + hex.EncodeToString(verificationBytes)

	// Create new claim
	claim, err := queries.CreateCharacterClaim(ctx, db.CreateCharacterClaimParams{
		CharacterID:      character.ID,
		ClaimerID:        userID,
		VerificationCode: verificationCode,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create claim")
	}

	resp := StartClaimResponse{
		ClaimID:          claim.ID.String(),
		VerificationCode: claim.VerificationCode,
		Status:           claim.Status,
		ClaimerID:        userID.String(),
	}

	return c.JSON(http.StatusCreated, resp)
}

// CheckClaim checks the status of a character claim
func (h *ClaimsHandler) CheckClaim(c echo.Context) error {
	log.Println("CheckClaim called")
	claimID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid claim ID")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user ID format")
	}

	queries := db.New(h.connPool)
	ctx := c.Request().Context()

	// Get claim by claim ID
	claim, err := queries.GetClaimByID(ctx, claimID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "claim not found")
	}

	// Verify the claim belongs to the user
	if claim.ClaimerID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "claim does not belong to user")
	}

	// If claim is pending, verify it
	if claim.Status == "pending" {
		// Check TibiaData API for verification code
		verified, err := h.tibiaData.VerifyCharacterClaim(claim.CharacterName, claim.VerificationCode)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify claim")
		}

		if verified {
			// Update claim status to approved
			updatedClaim, err := queries.UpdateClaimStatus(ctx, db.UpdateClaimStatusParams{
				CharacterID: claim.CharacterID,
				ClaimerID:   claim.ClaimerID,
				Status:      "approved",
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to update claim status")
			}

			// First deactivate any existing list memberships
			err = queries.DeactivateCharacterListMemberships(ctx, claim.CharacterID)
			if err != nil {
				log.Printf("Failed to deactivate list memberships for character %s: %v", claim.CharacterName, err)
				// Continue anyway as this is not critical
			}

			// Update character owner
			character, err := queries.UpdateCharacterOwner(ctx, db.UpdateCharacterOwnerParams{
				ID:     claim.CharacterID,
				UserID: claim.ClaimerID,
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to update character owner")
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"claim_id":  updatedClaim.ID,
				"status":    updatedClaim.Status,
				"character": character,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"claim_id":          claim.ID,
		"verification_code": claim.VerificationCode,
		"status":            claim.Status,
	})
}

// ProcessPendingClaims processes all pending claims that are due for check
func (h *ClaimsHandler) ProcessPendingClaims() error {
	log.Println("Processing pending claims")
	queries := db.New(h.connPool)
	ctx := context.Background()

	pendingClaims, err := queries.GetPendingClaimsToCheck(ctx)
	if err != nil {
		return err
	}

	for _, claim := range pendingClaims {
		// Get character
		character, err := queries.GetCharacter(ctx, claim.CharacterID)
		if err != nil {
			continue
		}

		// Check TibiaData API for verification code
		verified, err := h.tibiaData.VerifyCharacterClaim(character.Name, claim.VerificationCode)
		if err != nil {
			continue
		}

		status := "pending"
		if verified {
			status = "approved"
		} else if time.Since(claim.CreatedAt.Time) > 24*time.Hour {
			status = "rejected"
		}

		// Update claim status
		_, err = queries.UpdateClaimStatus(ctx, db.UpdateClaimStatusParams{
			CharacterID: claim.CharacterID,
			ClaimerID:   claim.ClaimerID,
			Status:      status,
		})
		if err != nil {
			continue
		}

		// If claim is approved, update character owner and deactivate list memberships
		if status == "approved" {
			// First deactivate any existing list memberships
			err = queries.DeactivateCharacterListMemberships(ctx, character.ID)
			if err != nil {
				log.Printf("Failed to deactivate list memberships for character %s: %v", character.Name, err)
				// Continue anyway as this is not critical
			}

			// Then update the character owner
			_, err = queries.UpdateCharacterOwner(ctx, db.UpdateCharacterOwnerParams{
				ID:     character.ID,
				UserID: claim.ClaimerID,
			})
			if err != nil {
				continue
			}
		}
	}

	return nil
}
