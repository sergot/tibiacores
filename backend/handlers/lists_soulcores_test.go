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
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddSoulcore(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  string
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
			expectedCode: "success",
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  "validation_error",
			expectedError: "Invalid list ID",
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
			expectedCode:  "validation_error",
			expectedError: "Invalid request body",
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
			expectedCode:  "authorization_error",
			expectedError: "User is not a member of this list",
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
			expectedCode:  "database_error",
			expectedError: "Failed to add soul core",
		}, {
			name: "Invalid Status Value",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				reqBody.Reset()
				// Need to access the current test's creatureID parameter here
				err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
					"creature_id": "invalid-status", // This will cause the handler to fail
					"status":      "invalid-status", // Invalid status value
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks expected as validation should fail before DB calls
			},
			expectedCode:  "validation_error",
			expectedError: "Invalid request body",
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
				require.Error(t, err)
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, appErr.Code)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, rec.Code)
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
		expectedCode  string
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
			expectedCode: "success",
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
						AuthorID: userID, // Current user is the owner
					}, nil)

				store.EXPECT().
					RemoveListSoulcore(gomock.Any(), db.RemoveListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(nil)
			},
			expectedCode: "success",
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid", creatureID.String())
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  "validation_error",
			expectedError: "Invalid list ID",
		},
		{
			name: "Invalid Creature ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues(listID.String(), "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  "validation_error",
			expectedError: "Invalid creature ID",
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
			expectedCode:  "authorization_error",
			expectedError: "User is not a member of this list",
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
			expectedCode:  "not_found_error",
			expectedError: "Soulcore not found",
		},
		{
			name: "Not Authorized to Remove",
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
						AuthorID: uuid.New(), // Different user is the owner
					}, nil)
			},
			expectedCode:  "authorization_error",
			expectedError: "Only the list owner or the user who added the soulcore can remove it",
		},
		{
			name: "Error Getting List",
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
						AddedByUserID: uuid.New(),
						Status:        db.SoulcoreStatusObtained,
					}, nil)

				store.EXPECT().
					GetList(gomock.Any(), listID).
					Return(db.List{}, errors.New("database error"))
			},
			expectedCode:  "database_error",
			expectedError: "Failed to get list details",
		},
		{
			name: "Database Error",
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
			expectedCode:  "database_error",
			expectedError: "Failed to remove soul core",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
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
				require.Error(t, err)
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, appErr.Code)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestUpdateSoulcoreStatus(t *testing.T) {
	listID := uuid.New()
	creatureID := uuid.New()

	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, reqBody *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID)
		expectedCode  string
		expectedError string
	}{
		{
			name: "Success - Soulcore Adder",
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

				// Expect call to get list members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:        userID, // The adder
							CharacterID:   uuid.New(),
							CharacterName: "TestCharacter",
							IsActive:      true,
						},
						{
							UserID:        uuid.New(), // Another user
							CharacterID:   uuid.New(),
							CharacterName: "OtherCharacter",
							IsActive:      true,
						},
					}, nil)

				// Expect adding directly to the adder's character
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), gomock.Any()).
					Return(nil)

				// Expect creating a suggestion for the other character
				store.EXPECT().
					CreateSoulcoreSuggestion(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: "success",
		},
		{
			name: "Success - List Owner",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Different user added the soulcore
				adderID := uuid.New()

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
						AddedByUserID: adderID, // Different user added the soulcore
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

				// Expect call to get list members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:        userID, // The owner
							CharacterID:   uuid.New(),
							CharacterName: "OwnerCharacter",
							IsActive:      true,
						},
						{
							UserID:        adderID, // The adder
							CharacterID:   uuid.New(),
							CharacterName: "AdderCharacter",
							IsActive:      true,
						},
					}, nil)

				// Expect attempting to add to both characters
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				store.EXPECT().
					CreateSoulcoreSuggestion(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedCode: "success",
		},
		{
			name: "Success - Update to Obtained Status",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
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
						AuthorID: uuid.New(),
					}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusObtained,
					}).
					Return(nil)
			},
			expectedCode: "success",
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  "validation_error",
			expectedError: "Invalid list ID",
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
			expectedCode:  "validation_error",
			expectedError: "Invalid request body",
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
			expectedCode:  "authorization_error",
			expectedError: "User is not a member of this list",
		},
		{
			name: "Soulcore Not Found",
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
					GetListSoulcore(gomock.Any(), db.GetListSoulcoreParams{
						ListID:     listID,
						CreatureID: creatureID,
					}).
					Return(db.GetListSoulcoreRow{}, sql.ErrNoRows)
			},
			expectedCode:  "not_found_error",
			expectedError: "Soulcore not found",
		},
		{
			name: "Not Authorized to Update",
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
						AuthorID: uuid.New(), // Different user is the owner
					}, nil)
			},
			expectedCode:  "authorization_error",
			expectedError: "Only the list owner or the user who added the soulcore can modify it",
		},
		{
			name: "Database Error",
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
			expectedCode:  "database_error",
			expectedError: "Failed to update soul core status",
		},
		{
			name: "Error Creating Suggestions",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Another member's ID
				otherMemberID := uuid.New()
				otherCharacterID := uuid.New()

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

				// Expect call to get list members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:        userID, // Current user (adder)
							CharacterID:   uuid.New(),
							CharacterName: "TestCharacter",
							IsActive:      true,
						},
						{
							UserID:        otherMemberID, // Another member
							CharacterID:   otherCharacterID,
							CharacterName: "OtherCharacter",
							IsActive:      true,
						},
					}, nil)

				// Expect adding soulcore to current user's character
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), gomock.Any()).
					Return(nil)

				// Expect creating a suggestion for the other user's character that returns an error
				store.EXPECT().
					CreateSoulcoreSuggestion(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedCode: "success", // Should still succeed even if suggestions creation fails
		},
		{
			name: "Unique Constraint Handling",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Another user ID
				otherMemberID := uuid.New()
				otherCharacterID := uuid.New()

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

				// Expect call to get list members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:        userID, // Current user (adder)
							CharacterID:   uuid.New(),
							CharacterName: "TestCharacter",
							IsActive:      true,
						},
						{
							UserID:        otherMemberID, // Another member
							CharacterID:   otherCharacterID,
							CharacterName: "OtherCharacter",
							IsActive:      true,
						},
					}, nil)

				// Simulate duplicate key error for adder's character
				duplicateError := errors.New("pq: duplicate key value violates unique constraint")
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), gomock.Any()).
					Return(duplicateError)

				// Simulate duplicate key error for other user's character
				store.EXPECT().
					CreateSoulcoreSuggestion(gomock.Any(), gomock.Any()).
					Return(duplicateError)
			},
			expectedCode: "success", // Should still succeed with duplicate key errors
		},
		{
			name: "Filter Inactive Members",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Basic setup for authentication and authorization
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
					Return(db.List{ID: listID}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				// Return mix of active and inactive members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{
						{
							UserID:        userID,
							CharacterID:   uuid.New(),
							CharacterName: "ActiveCharacter",
							IsActive:      true,
						},
						{
							UserID:        uuid.New(),
							CharacterID:   uuid.New(),
							CharacterName: "InactiveCharacter",
							IsActive:      false, // Inactive member
						},
					}, nil)

				// Expect only one call for the active member
				store.EXPECT().
					AddCharacterSoulcore(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
				// No calls expected for suggestion creation
			},
			expectedCode: "success",
		},
		{
			name: "GetListMembers Error Handling",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Basic setup for authentication and authorization
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
					Return(db.List{ID: listID}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				// Simulate database error when getting list members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return(nil, errors.New("database connection error"))
			},
			expectedCode: "success", // Should still succeed despite the error
		},
		{
			name: "Empty List Members",
			setupRequest: func(c echo.Context, reqBody *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, creatureID uuid.UUID, userID uuid.UUID) {
				// Basic setup
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
					Return(db.List{ID: listID}, nil)

				store.EXPECT().
					UpdateSoulcoreStatus(gomock.Any(), db.UpdateSoulcoreStatusParams{
						ListID:     listID,
						CreatureID: creatureID,
						Status:     db.SoulcoreStatusUnlocked,
					}).
					Return(nil)

				// Return empty list of members
				store.EXPECT().
					GetListMembersWithUnlocks(gomock.Any(), listID).
					Return([]db.GetListMembersWithUnlocksRow{}, nil)

				// No calls to AddCharacterSoulcore or CreateSoulcoreSuggestion expected
			},
			expectedCode: "success",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			userID := uuid.New()

			// Create request body
			reqBody := &bytes.Buffer{}
			err := json.NewEncoder(reqBody).Encode(map[string]interface{}{
				"creature_id": creatureID,
				"status":      db.SoulcoreStatusUnlocked,
			})
			require.NoError(t, err)

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/soulcores/%s/status", listID.String(), creatureID.String())
			req := httptest.NewRequest(http.MethodPut, url, reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/soulcores/:creature_id/status")
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
			err = h.UpdateSoulcoreStatus(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				appErr, ok := err.(*apperror.AppError)
				require.True(t, ok)
				require.Equal(t, tc.expectedCode, appErr.Code)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
