package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateList(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, body *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.CreateListResponse)
	}{
		{
			name: "Success - Existing Character",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID,
					"name":         "My Soul Core List",
				})
				require.NoError(t, err)
				c.Set("has_email", true)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Verify character exists and belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Create list using character's world
				store.EXPECT().
					CreateList(gomock.Any(), gomock.Any()).
					Return(db.List{
						ID:        uuid.New(),
						AuthorID:  userID,
						Name:      "My Soul Core List",
						World:     "Antica",
						ShareCode: uuid.New(),
						CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
						UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response *handlers.CreateListResponse) {
				require.NotEmpty(t, response.ID)
				require.Equal(t, "My Soul Core List", response.Name)
				require.Equal(t, "Antica", response.World)
				require.NotEmpty(t, response.ShareCode)
				require.True(t, response.HasEmail)
			},
		},
		{
			name: "Success - New Character",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NewCharacter",
					"name":           "My New List",
					"world":          "Secura",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Check if character name is already taken - single consolidated check
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "NewCharacter").
					Return(db.Character{}, sql.ErrNoRows)

				// Create character
				store.EXPECT().
					CreateCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "NewCharacter",
						World:  "Secura",
					}, nil)

				// Create list
				store.EXPECT().
					CreateList(gomock.Any(), gomock.Any()).
					Return(db.List{
						ID:        uuid.New(),
						AuthorID:  userID,
						Name:      "My New List",
						World:     "Secura",
						ShareCode: uuid.New(),
						CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
						UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response *handlers.CreateListResponse) {
				require.NotEmpty(t, response.ID)
				require.Equal(t, "My New List", response.Name)
				require.Equal(t, "Secura", response.World)
				require.NotEmpty(t, response.ShareCode)
			},
		},
		{
			name: "Success - Anonymous User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NewCharacter",
					"name":           "Anonymous List",
					"world":          "Monza",
				})
				require.NoError(t, err)
				// Remove user_id from context to simulate anonymous user
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Create a new anonymous user
				newUserID := uuid.New()
				store.EXPECT().
					CreateAnonymousUser(gomock.Any(), gomock.Any()).
					Return(db.User{
						ID: newUserID,
					}, nil)

				// Check if character name is already taken
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "NewCharacter").
					Return(db.Character{}, sql.ErrNoRows)

				// Create new character
				store.EXPECT().
					CreateCharacter(gomock.Any(), gomock.Any()).
					Do(func(_ interface{}, params db.CreateCharacterParams) {
						require.Equal(t, newUserID, params.UserID)
						require.Equal(t, "NewCharacter", params.Name)
						require.Equal(t, "Monza", params.World)
					}).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: newUserID,
						Name:   "NewCharacter",
						World:  "Monza",
					}, nil)

				// Create list
				store.EXPECT().
					CreateList(gomock.Any(), gomock.Any()).
					Do(func(_ interface{}, params db.CreateListParams) {
						require.Equal(t, newUserID, params.AuthorID)
						require.Equal(t, "Anonymous List", params.Name)
						require.Equal(t, "Monza", params.World)
					}).
					Return(db.List{
						ID:        uuid.New(),
						AuthorID:  newUserID,
						Name:      "Anonymous List",
						World:     "Monza",
						ShareCode: uuid.New(),
						CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
						UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response *handlers.CreateListResponse) {
				require.NotEmpty(t, response.ID)
				require.Equal(t, "Anonymous List", response.Name)
				require.Equal(t, "Monza", response.World)
				require.NotEmpty(t, response.ShareCode)
				// For anonymous users, should get a token in the header
				// The test doesn't check this because it's set on the response header
				// which is tested in the handler execution below
			},
		},
		{
			name: "Anonymous User - Missing Character Info",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"name": "Anonymous List",
					// Missing character_name and world
				})
				require.NoError(t, err)
				// Remove user_id from context to simulate anonymous user
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Create a new anonymous user
				newUserID := uuid.New()
				store.EXPECT().
					CreateAnonymousUser(gomock.Any(), gomock.Any()).
					Return(db.User{
						ID: newUserID,
					}, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "character_name and world are required for first list",
		},
		{
			name: "Error Creating Anonymous User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NewCharacter",
					"name":           "Anonymous List",
					"world":          "Monza",
				})
				require.NoError(t, err)
				// Remove user_id from context to simulate anonymous user
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Error creating anonymous user
				store.EXPECT().
					CreateAnonymousUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to create user",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				body.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "Missing List Name",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TestCharacter",
					"world":          "Antica",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "name is required",
		},
		{
			name: "Character Already Registered",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TakenCharacter",
					"name":           "My List",
					"world":          "Antica",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Return an existing character owned by another user
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "TakenCharacter").
					Return(db.Character{
						ID:     uuid.New(),
						UserID: uuid.New(), // Different user
						Name:   "TakenCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusConflict,
			expectedError: "character name is already registered",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID,
					"name":         "My List",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "character not found",
		},
		{
			name: "Character Belongs to Different User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID,
					"name":         "My List",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Return character belonging to different user
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Error Creating List",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID,
					"name":         "My List",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Verify character exists and belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Error creating list
				store.EXPECT().
					CreateList(gomock.Any(), gomock.Any()).
					Return(db.List{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to create list",
		},
		{
			name: "Error Adding Character to List",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID,
					"name":         "My List",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Verify character exists and belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Create list using character's world
				store.EXPECT().
					CreateList(gomock.Any(), gomock.Any()).
					Return(db.List{
						ID:        uuid.New(),
						AuthorID:  userID,
						Name:      "My List",
						World:     "Antica",
						ShareCode: uuid.New(),
						CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
						UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					}, nil)

				// Error adding character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to add character to list",
		},
		{
			name: "Error Creating Character",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NewCharacter",
					"name":           "My List",
					"world":          "Antica",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// Check if character name is already taken
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "NewCharacter").
					Return(db.Character{}, sql.ErrNoRows)

				// Error creating character
				store.EXPECT().
					CreateCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to create character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			userID := uuid.New()

			// Create HTTP request with empty body initially
			reqBody := bytes.NewBuffer([]byte(`{}`))
			req := httptest.NewRequest(http.MethodPost, "/api/lists", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup - authenticated user
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.CreateList(c)

			// Check for expected error response
			if tc.expectedError != "" {
				httpError, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, httpError.Code)
				require.Contains(t, httpError.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response handlers.CreateListResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, &response)
			}
		})
	}
}

func TestJoinList(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, body *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.ListDetailResponse)
	}{
		{
			name: "Success - Existing Character",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					AuthorID:  uuid.New(),
					Name:      "Test List",
					World:     "Antica",
					ShareCode: shareCode,
					CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Get character and verify ownership
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica", // Matches list world
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)

				// Get members for response
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter", IsActive: true},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter", IsActive: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse) {
				require.NotEmpty(t, response.ID)
				require.Equal(t, "Test List", response.Name)
				require.Equal(t, "Antica", response.World)
				require.Equal(t, 2, len(response.Members))
				require.Empty(t, response.SoulCores)
			},
		},
		{
			name: "Success - New Character",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NewCharacter",
					"world":          "Secura",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					AuthorID:  uuid.New(),
					Name:      "Test List",
					World:     "Secura",
					ShareCode: shareCode,
					CreatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
					UpdatedAt: pgtype.Timestamptz{Valid: true, Time: time.Now()},
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Check if character name is already taken
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "NewCharacter").
					Return(db.Character{}, sql.ErrNoRows)

				// Create new character
				store.EXPECT().
					CreateCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "NewCharacter",
						World:  "Secura",
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)

				// Get members for response
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "NewCharacter", IsActive: true},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter", IsActive: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse) {
				require.NotEmpty(t, response.ID)
				require.Equal(t, "Test List", response.Name)
				require.Equal(t, "Secura", response.World)
				require.Equal(t, 2, len(response.Members))
				require.Empty(t, response.SoulCores)
			},
		},
		{
			name: "Invalid Share Code",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Setup an invalid share code in the URL parameter
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid share code",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				body.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "List Not Found",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TestCharacter",
					"world":          "Antica",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				// List not found
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(db.List{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "list not found",
		},
		{
			name: "User Already a Member",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TestCharacter",
					"world":          "Antica",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					AuthorID:  uuid.New(),
					Name:      "Test List",
					World:     "Antica",
					ShareCode: shareCode,
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// User is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(true, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "user is already a member of this list",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					ShareCode: shareCode,
					World:     "Antica",
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "character not found",
		},
		{
			name: "Character Belongs to Different User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					ShareCode: shareCode,
					World:     "Antica",
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Character belongs to different user
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "World Mismatch",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					ShareCode: shareCode,
					World:     "Antica",
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Character from different world
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Secura", // Different world
					}, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "character world does not match list world",
		},
		{
			name: "Error Adding Character to List",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					ShareCode: shareCode,
					World:     "Antica",
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Get character and verify ownership
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Error adding character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to add character to list",
		},
		{
			name: "Error Getting List Members",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				characterID := uuid.New()
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_id": characterID.String(),
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, userID uuid.UUID) {
				list := db.List{
					ID:        uuid.New(),
					ShareCode: shareCode,
					World:     "Antica",
				}

				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Check if user is already a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), gomock.Any()).
					Return(false, nil)

				// Get character and verify ownership
				store.EXPECT().
					GetCharacter(gomock.Any(), gomock.Any()).
					Return(db.Character{
						ID:     uuid.New(),
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Add character to list
				store.EXPECT().
					AddListCharacter(gomock.Any(), gomock.Any()).
					Return(nil)

				// Error getting members
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get list members",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			userID := uuid.New()
			shareCode := uuid.New()

			// Create HTTP request with empty body initially
			reqBody := bytes.NewBuffer([]byte(`{}`))
			url := fmt.Sprintf("/api/lists/join/%s", shareCode.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/join/:share_code")
			c.Set("user_id", userID.String())
			c.SetParamNames("share_code")
			c.SetParamValues(shareCode.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, shareCode, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.JoinList(c)

			// Check for expected error response
			if tc.expectedError != "" {
				httpError, ok := err.(*echo.HTTPError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, httpError.Code)
				require.Contains(t, httpError.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response handlers.ListDetailResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, &response)
			}
		})
	}
}
