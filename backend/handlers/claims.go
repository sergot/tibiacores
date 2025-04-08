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
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/errors"
	"github.com/sergot/tibiacores/backend/services"
)

type ClaimsHandler struct {
	store     db.Store
	TibiaData services.TibiaDataServiceInterface
}

type StartClaimResponse struct {
	ClaimID          string `json:"claim_id"`
	VerificationCode string `json:"verification_code"`
	Status           string `json:"status"`
	ClaimerID        string `json:"claimer_id,omitempty"` // ID of the claiming user
}

func NewClaimsHandler(store db.Store) *ClaimsHandler {
	return &ClaimsHandler{
		store:     store,
		TibiaData: services.NewTibiaDataService(),
	}
}

// StartClaim initiates a character claim process
func (h *ClaimsHandler) StartClaim(c echo.Context) error {
	var req struct {
		CharacterName string `json:"character_name"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return errors.NewInvalidRequestError(err).
			WithOperation("decode_request").
			WithResource("claim")
	}

	if req.CharacterName == "" {
		return errors.NewInvalidRequestError(errors.ErrInvalidRequest).
			WithOperation("validate_request").
			WithResource("claim")
	}

	ctx := c.Request().Context()

	// First check if character exists in TibiaData API
	tibiaChar, err := h.TibiaData.GetCharacter(req.CharacterName)
	if err != nil {
		return errors.NewNotFoundError(err).
			WithOperation("get_tibia_character").
			WithResource("character")
	}

	// Check if character exists in our database
	character, err := h.store.GetCharacterByName(ctx, tibiaChar.Name)
	if err != nil {
		return errors.NewNotFoundError(err).
			WithOperation("get_character").
			WithResource("character")
	}

	// Get or create user
	var userID uuid.UUID
	var token string
	// Check if user is authenticated
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return errors.NewUnauthorizedError(err).
				WithOperation("parse_user_id").
				WithResource("user")
		}
	} else {
		// Create new anonymous user account
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return errors.NewDatabaseError(err).
				WithOperation("create_anonymous_user").
				WithResource("user")
		}
		userID = newUser.ID

		// Generate token
		token, err = auth.GenerateToken(userID.String(), false)
		if err != nil {
			return errors.NewInternalError(err).
				WithOperation("generate_token").
				WithResource("auth")
		}

		c.Response().Header().Set("X-Auth-Token", token)
	}

	// Check if there's already an active claim
	existingClaim, err := h.store.GetCharacterClaim(ctx, db.GetCharacterClaimParams{
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
		return errors.NewInternalError(err).
			WithOperation("generate_verification_code").
			WithResource("claim")
	}
	verificationCode := "TIBIACORES-" + hex.EncodeToString(verificationBytes)

	// Create new claim
	claim, err := h.store.CreateCharacterClaim(ctx, db.CreateCharacterClaimParams{
		CharacterID:      character.ID,
		ClaimerID:        userID,
		VerificationCode: verificationCode,
	})
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("create_claim").
			WithResource("claim")
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
	claimID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return errors.NewInvalidRequestError(err).
			WithOperation("parse_claim_id").
			WithResource("claim")
	}

	// Get authenticated user ID from context
	userIDStr := c.Get("user_id").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.NewUnauthorizedError(err).
			WithOperation("parse_user_id").
			WithResource("user")
	}

	ctx := c.Request().Context()

	// Get claim by claim ID
	claim, err := h.store.GetClaimByID(ctx, claimID)
	if err != nil {
		return errors.NewNotFoundError(err).
			WithOperation("get_claim").
			WithResource("claim")
	}

	// Verify the claim belongs to the user
	if claim.ClaimerID != userID {
		return errors.NewForbiddenError(errors.ErrForbidden).
			WithOperation("verify_claim_ownership").
			WithResource("claim")
	}

	// If claim is pending, verify it
	if claim.Status == "pending" {
		// Check TibiaData API for verification code
		verified, err := h.TibiaData.VerifyCharacterClaim(claim.CharacterName, claim.VerificationCode)
		if err != nil {
			return errors.NewInternalError(err).
				WithOperation("verify_claim").
				WithResource("claim")
		}

		if verified {
			// Update claim status to approved
			updatedClaim, err := h.store.UpdateClaimStatus(ctx, db.UpdateClaimStatusParams{
				CharacterID: claim.CharacterID,
				ClaimerID:   claim.ClaimerID,
				Status:      "approved",
			})
			if err != nil {
				return errors.NewDatabaseError(err).
					WithOperation("update_claim_status").
					WithResource("claim")
			}

			// First deactivate any existing list memberships
			err = h.store.DeactivateCharacterListMemberships(ctx, claim.CharacterID)
			if err != nil {
				log.Printf("Failed to deactivate list memberships for character %s: %v", claim.CharacterName, err)
				// Continue anyway as this is not critical
			}

			// Update character owner
			character, err := h.store.UpdateCharacterOwner(ctx, db.UpdateCharacterOwnerParams{
				ID:     claim.CharacterID,
				UserID: claim.ClaimerID,
			})
			if err != nil {
				return errors.NewDatabaseError(err).
					WithOperation("update_character_owner").
					WithResource("character")
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
	ctx := context.Background()

	pendingClaims, err := h.store.GetPendingClaimsToCheck(ctx)
	if err != nil {
		return errors.NewDatabaseError(err).
			WithOperation("get_pending_claims").
			WithResource("claim")
	}

	for _, claim := range pendingClaims {
		// Get character
		character, err := h.store.GetCharacter(ctx, claim.CharacterID)
		if err != nil {
			log.Printf("Failed to get character for claim %s: %v", claim.ID, err)
			continue
		}

		// Check TibiaData API for verification code
		verified, err := h.TibiaData.VerifyCharacterClaim(character.Name, claim.VerificationCode)
		if err != nil {
			log.Printf("Failed to verify claim %s with TibiaData API: %v", claim.ID, err)
			continue
		}

		status := "pending"
		if verified {
			status = "approved"
		} else if time.Since(claim.CreatedAt.Time) > 24*time.Hour {
			status = "rejected"
		}

		// Update claim status
		_, err = h.store.UpdateClaimStatus(ctx, db.UpdateClaimStatusParams{
			CharacterID: claim.CharacterID,
			ClaimerID:   claim.ClaimerID,
			Status:      status,
		})
		if err != nil {
			log.Printf("Failed to update claim status for claim %s: %v", claim.ID, err)
			continue
		}
	}

	return nil
}
