package handlers_test

import (
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetListMembersWithUnlocks(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []db.GetListMembersWithUnlocksRow)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				// Check if user is a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				// Get members with unlocks
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:            userID,
							CharacterID:       uuid.New(),
							CharacterName:     "TestCharacter",
							UnlockedCreatures: json.RawMessage(`[{"creature_id":"123","creature_name":"Dragon"}]`),
							ObtainedCount:     5,
							UnlockedCount:     2,
							IsActive:          true,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []db.GetListMembersWithUnlocksRow) {
				require.Len(t, response, 1)
				require.Equal(t, "TestCharacter", response[0].CharacterName)
				require.Equal(t, int64(5), response[0].ObtainedCount)
				require.Equal(t, int64(2), response[0].UnlockedCount)
				require.True(t, response[0].IsActive)
			},
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "Invalid User ID Format",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "invalid user ID format",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				// Check if user is a member - returns false
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
			name: "Database Error Getting Members",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				// Check if user is a member
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				// Database error when getting members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
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
			listID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/members", listID.String())
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/members")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(listID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetListMembersWithUnlocks(c)

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
				var response []db.GetListMembersWithUnlocksRow
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}