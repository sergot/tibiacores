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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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

func TestRemoveSoulcore(t *testing.T) {
	listID := uuid.New()
	creatureID := uuid.New()

	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Soulcore Adder",
			setupRequest: func(c echo.Context) {
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
					RemoveListSoulcore(gomock.Any(), db.RemoveListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success - List Owner",
			setupRequest: func(c echo.Context) {
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
						AuthorID: userID, // Current user is the list owner
					}, nil)

				store.EXPECT().
					RemoveListSoulcore(gomock.Any(), db.RemoveListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid", creatureID.String())
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid list ID",
		},
		{
			name: "Invalid Creature ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues(listID.String(), "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid creature ID",
		},
		{
			name: "User Not a Member",
			setupRequest: func(c echo.Context) {
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
			name: "Soulcore Not Found",
			setupRequest: func(c echo.Context) {
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
			setupRequest: func(c echo.Context) {
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
			expectedError: "only the list owner or the user who added the soulcore can remove it",
		},
		{
			name: "Error Removing Soulcore",
			setupRequest: func(c echo.Context) {
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
					RemoveListSoulcore(gomock.Any(), db.RemoveListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to remove soul core",
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

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/soulcores/%s", listID.String(), creatureID.String())
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/soulcores/:creature_id")
			c.Set("user_id", userID.String())
			c.SetParamNames("id", "creature_id")
			c.SetParamValues(listID.String(), creatureID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, creatureID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.RemoveSoulcore(c)

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
