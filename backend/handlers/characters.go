package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
)

type CharactersHandler struct {
	store db.Store
}

func NewCharactersHandler(store db.Store) *CharactersHandler {
	return &CharactersHandler{
		store: store,
	}
}

type HighscoreResponse struct {
	Characters []CharacterScore `json:"characters"`
	Pagination PaginationInfo   `json:"pagination"`
}

type CharacterScore struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	World     string `json:"world"`
	CoreCount int64  `json:"core_count"`
}

type PaginationInfo struct {
	TotalPages   int `json:"total_pages"`
	CurrentPage  int `json:"current_page"`
	TotalRecords int `json:"total_records"`
	PageSize     int `json:"page_size"`
}

func (h *CharactersHandler) GetHighscores(c echo.Context) error {
	// Get page number from query parameters, default to 1
	pageStr := c.QueryParam("page")
	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return apperror.ValidationError("Invalid page number", err).
				WithDetails(&apperror.ValidationErrorDetails{
					Field:  "page",
					Value:  pageStr,
					Reason: "Page must be a positive integer",
				})
		}
	}

	// Constants for pagination
	const pageSize = 20
	const maxPages = 50

	// Validate page number against max pages
	if page > maxPages {
		return apperror.ValidationError("Page number too high", nil).
			WithDetails(&apperror.ValidationErrorDetails{
				Field:  "page",
				Value:  strconv.Itoa(page),
				Reason: "Maximum page number is 50",
			})
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get characters sorted by core count with pagination
	ctx := c.Request().Context()
	characters, err := h.store.GetHighscoreCharacters(ctx, db.GetHighscoreCharactersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return apperror.DatabaseError("Failed to get highscore characters", err).
			WithDetails(&apperror.DatabaseErrorDetails{
				Operation: "GetHighscoreCharacters",
				Table:     "characters",
			}).
			Wrap(err)
	}

	// Calculate pagination info
	var totalRecords int64
	if len(characters) > 0 {
		totalRecords = characters[0].TotalCount
	}
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	// Ensure we have at least 1 page even with no results
	if totalPages == 0 {
		totalPages = 1
	}

	// Map to response format
	characterScores := make([]CharacterScore, len(characters))
	for i, c := range characters {
		characterScores[i] = CharacterScore{
			ID:        c.ID.String(),
			Name:      c.Name,
			World:     c.World,
			CoreCount: c.CoreCount,
		}
	}

	return c.JSON(http.StatusOK, HighscoreResponse{
		Characters: characterScores,
		Pagination: PaginationInfo{
			TotalPages:   totalPages,
			CurrentPage:  page,
			TotalRecords: int(totalRecords),
			PageSize:     pageSize,
		},
	})
}
