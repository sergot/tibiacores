package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
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
		return apperror.ValidationError("Invalid request body", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "body",
				Reason: "Invalid JSON format",
			})
	}

	ctx := c.Request().Context()

	// First check if character exists in TibiaData API
	tibiaChar, err := h.TibiaData.GetCharacter(req.CharacterName)
	if err != nil {
		return apperror.NotFoundError("Character not found in Tibia", err).
			WithDetails(&apperror.ExternalServiceErrorDetails{
				Service:   "TibiaData",
				Operation: "GetCharacter",
			}).
			Wrap(err)
	}

	// Check if character exists in our database
	character, err := h.store.GetCharacterByName(ctx, tibiaChar.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Character is not registered in any list yet", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetCharacterByName",
					Table:     "characters",
				})
		}
		return apperror.DatabaseError("Failed to check character in database", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetCharacterByName",
				Table:     "characters",
			}).
			Wrap(err)
	}

	// Get or create user
	var userID uuid.UUID
	var token string
	// Check if user is authenticated
	if userIDStr, ok := c.Get("user_id").(string); ok && userIDStr != "" {
		// User is authenticated, parse their ID
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			return apperror.AuthorizationError("Invalid user ID format", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "user_id",
					Value:  userIDStr,
					Reason: "Invalid UUID format",
				})
		}
	} else {
		// Create new anonymous user account
		newUser, err := h.store.CreateAnonymousUser(ctx, uuid.New())
		if err != nil {
			return apperror.DatabaseError("Failed to create user", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "CreateAnonymousUser",
					Table:     "users",
				}).
				Wrap(err)
		}
		userID = newUser.ID

		// Generate token
		token, err = auth.GenerateToken(userID.String(), false)
		if err != nil {
			return apperror.InternalError("Failed to generate authentication token", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "token",
					Reason: "Token generation failed",
				}).
				Wrap(err)
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
		return apperror.InternalError("Failed to generate verification code", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "verification_code",
				Reason: "Random number generation failed",
			}).
			Wrap(err)
	}
	verificationCode := "TIBIACORES-" + hex.EncodeToString(verificationBytes)

	// Create new claim
	claim, err := h.store.CreateCharacterClaim(ctx, db.CreateCharacterClaimParams{
		CharacterID:      character.ID,
		ClaimerID:        userID,
		VerificationCode: verificationCode,
	})
	if err != nil {
		return apperror.DatabaseError("Failed to create claim", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "CreateCharacterClaim",
				Table:     "character_claims",
			}).
			Wrap(err)
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
		return apperror.ValidationError("Invalid claim ID", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "claim_id",
				Value:  c.Param("id"),
				Reason: "Invalid UUID format",
			})
	}

	// Get authenticated user ID from context
	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		return apperror.AuthorizationError("Missing user authentication", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Reason: "Not found in context",
			})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperror.AuthorizationError("Invalid user ID format", err).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "user_id",
				Value:  userIDStr,
				Reason: "Invalid UUID format",
			})
	}

	ctx := c.Request().Context()

	// Get claim by claim ID
	claim, err := h.store.GetClaimByID(ctx, claimID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.NotFoundError("Claim not found", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetClaimByID",
					Table:     "character_claims",
				})
		}
		return apperror.DatabaseError("Failed to retrieve claim", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetClaimByID",
				Table:     "character_claims",
			}).
			Wrap(err)
	}

	// Verify the claim belongs to the user
	if claim.ClaimerID != userID {
		return apperror.AuthorizationError("Claim does not belong to this user", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "claimer_id",
				Value:  userID.String(),
				Reason: "Claim belongs to different user",
			})
	}

	// If claim is pending, verify it
	if claim.Status == "pending" {
		// Check TibiaData API for verification code
		verified, err := h.TibiaData.VerifyCharacterClaim(claim.CharacterName, claim.VerificationCode)
		if err != nil {
			return apperror.ExternalServiceError("Failed to verify claim with Tibia Data service", err).
				WithDetails(&apperror.ExternalServiceErrorDetails{
					Service:   "TibiaData",
					Operation: "VerifyCharacterClaim",
				}).
				Wrap(err)
		}

		if verified {
			// Update claim status to approved
			updatedClaim, err := h.store.UpdateClaimStatus(ctx, db.UpdateClaimStatusParams{
				CharacterID: claim.CharacterID,
				ClaimerID:   claim.ClaimerID,
				Status:      "approved",
			})
			if err != nil {
				return apperror.DatabaseError("Failed to update claim status", err).
					WithDetails(&apperror.DatabaseErrorDetails{
						Operation: "UpdateClaimStatus",
						Table:     "character_claims",
					}).
					Wrap(err)
			}

			// First deactivate any existing list memberships
			err = h.store.DeactivateCharacterListMemberships(ctx, claim.CharacterID)
			if err != nil {
				// Log the error but continue as this is not critical
				apperror.DatabaseError("Failed to deactivate list memberships", err).
					WithDetails(&apperror.DatabaseErrorDetails{
						Operation: "DeactivateCharacterListMemberships",
						Table:     "list_characters",
					}).
					LogError()
			}

			// Update character owner
			character, err := h.store.UpdateCharacterOwner(ctx, db.UpdateCharacterOwnerParams{
				ID:     claim.CharacterID,
				UserID: claim.ClaimerID,
			})
			if err != nil {
				return apperror.DatabaseError("Failed to update character owner", err).
					WithDetails(&apperror.DatabaseErrorDetails{
						Operation: "UpdateCharacterOwner",
						Table:     "characters",
					}).
					Wrap(err)
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
		return apperror.DatabaseError("Failed to fetch pending claims", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetPendingClaimsToCheck",
				Table:     "character_claims",
			}).
			Wrap(err)
	}

	for _, claim := range pendingClaims {
		// Get character
		character, err := h.store.GetCharacter(ctx, claim.CharacterID)
		if err != nil {
			apperror.DatabaseError("Failed to get character for claim", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "GetCharacter",
					Table:     "characters",
				}).
				WithContext(apperror.ErrorContext{
					Operation: "ProcessPendingClaims",
				}).
				LogError()
			continue
		}

		// Check TibiaData API for verification code
		verified, err := h.TibiaData.VerifyCharacterClaim(character.Name, claim.VerificationCode)
		if err != nil {
			apperror.ExternalServiceError("Failed to verify claim with external service", err).
				WithDetails(&apperror.ExternalServiceErrorDetails{
					Service:   "TibiaData",
					Operation: "VerifyCharacterClaim",
				}).
				WithContext(apperror.ErrorContext{
					Operation: "ProcessPendingClaims",
				}).
				LogError()
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
			apperror.DatabaseError("Failed to update claim status", err).
				WithDetails(&apperror.DatabaseErrorDetails{
					Operation: "UpdateClaimStatus",
					Table:     "character_claims",
				}).
				WithContext(apperror.ErrorContext{
					Operation: "ProcessPendingClaims",
				}).
				LogError()
			continue
		}

		if status == "approved" {
			// Update character owner
			_, err = h.store.UpdateCharacterOwner(ctx, db.UpdateCharacterOwnerParams{
				ID:     claim.CharacterID,
				UserID: claim.ClaimerID,
			})
			if err != nil {
				apperror.DatabaseError("Failed to update character owner", err).
					WithDetails(&apperror.DatabaseErrorDetails{
						Operation: "UpdateCharacterOwner",
						Table:     "characters",
					}).
					WithContext(apperror.ErrorContext{
						Operation: "ProcessPendingClaims",
					}).
					LogError()
			}
		}
	}

	return nil
}
