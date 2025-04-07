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

func TestGetList(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, list db.List, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.ListDetailResponse, list db.List)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - includes current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter"},
					}, nil)

				// Get list soulcores
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return([]db.GetListSoulcoresRow{
						{
							CreatureID: uuid.New(),
							ListID:     list.ID,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, list.Name, response.Name)
				require.Equal(t, list.World, response.World)
				require.Equal(t, 2, len(response.Members))
				require.Equal(t, 1, len(response.SoulCores))
			},
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - doesn't include current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: uuid.New(), CharacterName: "OtherCharacter"},
						{UserID: list.AuthorID, CharacterName: "AuthorCharacter"},
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "user is not a member of this list",
		},
		{
			name: "List Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// List not found
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(db.List{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "list not found",
		},
		{
			name: "Error Getting List Members",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Error getting list members
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get list members",
		},
		{
			name: "Error Getting Soul Cores",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

				// Get list members - includes current user
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
					}, nil)

				// Error getting soulcores
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get soul cores",
		},
		{
			name: "Empty List (No Members or Soulcores)",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, list db.List, userID uuid.UUID) {
				// Get list details
				store.EXPECT().
					GetList(gomock.Any(), list.ID).
					Return(list, nil)

					// Only the current user is a member
					// Only the current user is a member
				store.EXPECT().
					GetListMembers(gomock.Any(), list.ID).
					Return([]db.GetListMembersRow{
						{UserID: userID, CharacterName: "TestCharacter"},
					}, nil)

				// Empty soulcores list
				store.EXPECT().
					GetListSoulcores(gomock.Any(), list.ID).
					Return([]db.GetListSoulcoresRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListDetailResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, 1, len(response.Members))
				require.Equal(t, 0, len(response.SoulCores))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			userID := uuid.New()
			list := db.List{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				Name:      "Test List",
				World:     "Antica",
				ShareCode: uuid.New(),
			}

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s", list.ID.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(list.ID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, list, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetList(c)

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
				tc.checkResponse(t, &response, list)
			}
		})
	}
}

func TestGetListPreview(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.ListPreviewResponse, list db.List)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List) {
				// Get list by share code
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				// Get members count
				store.EXPECT().
					GetMembers(gomock.Any(), list.ID).
					Return([]db.ListsUser{
						{UserID: uuid.New(), CharacterID: uuid.New(), ListID: list.ID, Active: true},
						{UserID: uuid.New(), CharacterID: uuid.New(), ListID: list.ID, Active: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListPreviewResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, list.Name, response.Name)
				require.Equal(t, list.World, response.World)
				require.Equal(t, 2, response.MemberCount)
			},
		},
		{
			name: "Invalid Share Code",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid share code",
		},
		{
			name: "List Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List) {
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(db.List{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "list not found",
		},
		{
			name: "Error Getting Members",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List) {
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				store.EXPECT().
					GetMembers(gomock.Any(), list.ID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get list members",
		},
		{
			name: "Empty List (No Members)",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, shareCode uuid.UUID, list db.List) {
				store.EXPECT().
					GetListByShareCode(gomock.Any(), shareCode).
					Return(list, nil)

				store.EXPECT().
					GetMembers(gomock.Any(), list.ID).
					Return([]db.ListsUser{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response *handlers.ListPreviewResponse, list db.List) {
				require.Equal(t, list.ID, response.ID)
				require.Equal(t, list.Name, response.Name)
				require.Equal(t, list.World, response.World)
				require.Equal(t, 0, response.MemberCount)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			shareCode := uuid.New()
			list := db.List{
				ID:        uuid.New(),
				Name:      "Test List",
				World:     "Antica",
				ShareCode: shareCode,
			}

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/preview/%s", shareCode.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/preview/:share_code")
			c.SetParamNames("share_code")
			c.SetParamValues(shareCode.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, shareCode, list)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetListPreview(c)

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
				var response handlers.ListPreviewResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, &response, list)
			}
		})
	}
}

func TestGetCharacterSuggestions(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetCharacterSuggestionsRow)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Get character suggestions - expect to be called exactly once
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return([]db.GetCharacterSuggestionsRow{
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							ListID:       uuid.New(),
							CreatureName: "Dragon",
						},
						{
							CharacterID:  characterID,
							CreatureID:   uuid.New(),
							ListID:       uuid.New(),
							CreatureName: "Demon",
						},
					}, nil).
					Times(1)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSuggestionsRow) {
				require.Equal(t, 2, len(response))
				require.Equal(t, "Dragon", response[0].CreatureName)
				require.Equal(t, "Demon", response[1].CreatureName)
			},
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid character ID",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Database Error Getting Suggestions",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Database error getting suggestions
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get suggestions",
		},
		{
			name: "No Suggestions",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Empty suggestions list
				store.EXPECT().
					GetCharacterSuggestions(gomock.Any(), characterID).
					Return([]db.GetCharacterSuggestionsRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetCharacterSuggestionsRow) {
				require.Empty(t, response)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			characterID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions", characterID.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetCharacterSuggestions(c)

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
				var response []db.GetCharacterSuggestionsRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestAcceptSoulcoreSuggestion(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Add soulcore to character - expect to be called exactly once
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)

				// Delete the suggestion - expect to be called exactly once
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid character ID",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Error Adding Soulcore",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Error adding soulcore
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to add soulcore to character",
		},
		{
			name: "Error Deleting Suggestion",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Add soulcore successfully
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), db.AddCharacterSoulcoreParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil)

				// Error deleting suggestion
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to delete suggestion",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			characterID := uuid.New()
			creatureID := uuid.New()
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
			})
			require.NoError(t, err)

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions/accept", characterID.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions/accept")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err = h.AcceptSoulcoreSuggestion(c)

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
		})
	}
}

func TestDismissSoulcoreSuggestion(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user - expect to be called exactly once
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil).
					Times(1)

				// Delete the suggestion - expect to be called exactly once
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(nil).
					Times(1)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid character ID",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Character not found
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get character",
		},
		{
			name: "Character Belongs To Different User",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Return a character with different userID
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "OtherCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "character does not belong to user",
		},
		{
			name: "Error Deleting Suggestion",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, characterID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Verify character belongs to user
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Error deleting suggestion
				store.EXPECT().
					DeleteSoulcoreSuggestion(gomock.Any(), db.DeleteSoulcoreSuggestionParams{
						CharacterID: characterID,
						CreatureID:  creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to delete suggestion",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			characterID := uuid.New()
			creatureID := uuid.New()
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
			})
			require.NoError(t, err)

			// Create HTTP request
			url := fmt.Sprintf("/api/characters/%s/suggestions/dismiss", characterID.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/characters/:id/suggestions/dismiss")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(characterID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, characterID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err = h.DismissSoulcoreSuggestion(c)

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
		})
	}
}

func TestAddSoulcore(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Check if user is a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				// Add soulcore to list
				store.EXPECT().
					AddSoulcoreToList(gomock.Any(), db.AddSoulcoreToListParams{
						ListID:        listID,
						CreatureID:    creatureID,
						Status:        db.SoulcoreStatusObtained,
						AddedByUserID: userID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				reqBody.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "user is not a member of this list",
		},
		{
			name: "Error Adding Soulcore",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					AddSoulcoreToList(gomock.Any(), db.AddSoulcoreToListParams{
						ListID:        listID,
						CreatureID:    creatureID,
						Status:        db.SoulcoreStatusObtained,
						AddedByUserID: userID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to add soul core",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			listID := uuid.New()
			creatureID := uuid.New()
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
				"status":      db.SoulcoreStatusObtained,
			})
			require.NoError(t, err)

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/soulcores", listID.String())
			req := httptest.NewRequest(http.MethodPost, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/soulcores")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(listID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err = h.AddSoulcore(c)

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
		})
	}
}

func TestUpdateSoulcoreStatus(t *testing.T) {
	// Create common IDs for all tests upfront
	sharedCreatureID := uuid.New()

	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Soulcore Adder",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: userID,
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: uuid.New(), // Different user is the owner
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				store.EXPECT().
					CreateSoulcoreSuggestions(gomock.Any(), db.CreateSoulcoreSuggestionsParams{
						ID:         listID,
						CreatureID: creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success - List Owner",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: uuid.New(), // Different user added the soulcore
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: userID, // Current user is the owner
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				store.EXPECT().
					CreateSoulcoreSuggestions(gomock.Any(), db.CreateSoulcoreSuggestionsParams{
						ID:         listID,
						CreatureID: creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success - Update to Obtained Status",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusObtained,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: userID,
						Status:        db.SoulcoreStatusUnlocked,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: uuid.New(), // Different user is the owner
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusObtained,
					}).
					Return(nil)
				// No suggestion creation call expected for obtained status
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				c.SetParamValues("invalid-uuid")

				// Still need valid JSON in the body
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				reqBody.Reset()
				reqBody.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				// Default setup
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "user is not a member of this list",
		},
		{
			name: "Soulcore Not Found",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				// Default setup
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "soulcore not found",
		},
		{
			name: "Neither Owner Nor Adder",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				// Default setup
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: uuid.New(), // Different user
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: uuid.New(), // Different user
					}, nil)
			},
			expectedCode:  http.StatusForbidden,
			expectedError: "only the list owner or the user who added the soulcore can modify it",
		},
		{
			name: "Error Updating Status",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				// Default setup
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: userID,
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: uuid.New(),
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to update soul core status",
		},
		{
			name: "Error Creating Suggestions",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer, creatureID uuid.UUID) {
				// Default setup
				reqBody.Reset()
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": creatureID,
					"status":      db.SoulcoreStatusUnlocked,
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{
						ListID:        listID,
						CreatureID:    creatureID,
						AddedByUserID: userID,
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{
						ID:       listID,
						AuthorID: uuid.New(),
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				store.EXPECT().
					CreateSoulcoreSuggestions(gomock.Any(), db.CreateSoulcoreSuggestionsParams{
						ID:         listID,
						CreatureID: creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode: http.StatusOK, // Should still succeed as suggestion errors are logged but not returned
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			listID := uuid.New()
			userID := uuid.New()

			// Create request body with a shared creatureID
			reqBody := &bytes.Buffer{}

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/soulcores", listID.String())
			req := httptest.NewRequest(http.MethodPut, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/soulcores")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(listID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody, sharedCreatureID)
			}

			// Setup mock expectations with the shared creatureID
			tc.setupMocks(store, listID, sharedCreatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.UpdateSoulcoreStatus(c)

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
		})
	}
}
