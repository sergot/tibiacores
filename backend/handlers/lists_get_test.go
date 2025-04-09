package handlers_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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
			expectedError: "Invalid list ID",
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
			expectedCode:  http.StatusUnauthorized,
			expectedError: "User is not a member of this list",
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
			expectedError: "List not found",
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
			expectedError: "Failed to get list members",
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
			expectedError: "Failed to get soul cores",
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
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, appErr.StatusCode)
				require.Contains(t, appErr.Message, tc.expectedError)
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
			expectedError: "Invalid share code",
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
			expectedError: "List not found",
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
			expectedError: "Failed to get list members",
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
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, appErr.StatusCode)
				require.Contains(t, appErr.Message, tc.expectedError)
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
