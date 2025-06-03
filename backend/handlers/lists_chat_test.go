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
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/sergot/tibiacores/backend/pkg/validator"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateChatMessage(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, body *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response handlers.ChatMessage)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				// Check if user is a member of the list
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				// Get character
				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				// Create chat message
				store.EXPECT().
					CreateChatMessage(gomock.Any(), db.CreateChatMessageParams{
						ListID:      listID,
						UserID:      userID,
						CharacterID: characterID,
						Message:     "Test message",
					}).
					Return(db.ListChatMessage{
						ID:          uuid.New(),
						ListID:      listID,
						UserID:      userID,
						CharacterID: characterID,
						Message:     "Test message",
						CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response handlers.ChatMessage) {
				require.Equal(t, "Test message", response.Message)
				require.Equal(t, "TestCharacter", response.CharacterName)
				require.NotEmpty(t, response.ID)
			},
		},
		{
			name: "Invalid List ID",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid list ID",
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "User Not List Member",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				// Check if user is a member of the list
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "User is not a member of this list",
		},
		{
			name: "Database Error - IsUserListMember",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to check list membership",
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				body.WriteString("invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Missing Message",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]any{
					"character_id": uuid.New().String(),
					// Missing message
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Validation failed",
		},
		{
			name: "Invalid Character ID",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]any{
					"message":      "Test message",
					"character_id": "invalid-uuid",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Validation failed",
		},
		{
			name: "Character Not Found",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get character",
		},
		{
			name: "Character Belongs to Different User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: uuid.New(), // Different user
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Character does not belong to user",
		},
		{
			name: "Database Error - CreateChatMessage",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID, characterID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetCharacter(gomock.Any(), characterID).
					Return(db.Character{
						ID:     characterID,
						UserID: userID,
						Name:   "TestCharacter",
						World:  "Antica",
					}, nil)

				store.EXPECT().
					CreateChatMessage(gomock.Any(), db.CreateChatMessageParams{
						ListID:      listID,
						UserID:      userID,
						CharacterID: characterID,
						Message:     "Test message",
					}).
					Return(db.ListChatMessage{}, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to create chat message",
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
			characterID := uuid.New()

			// Create HTTP request
			reqBody := bytes.NewBuffer([]byte(`{}`))
			req := httptest.NewRequest(http.MethodPost, "/api/lists/"+listID.String()+"/chat", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = validator.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/chat")
			c.SetParamNames("id")
			c.SetParamValues(listID.String())
			c.Set("user_id", userID.String())

			// Setup default request body
			reqBody.Reset()
			err := json.NewEncoder(reqBody).Encode(map[string]any{
				"message":      "Test message",
				"character_id": characterID.String(),
			})
			require.NoError(t, err)

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, userID, characterID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err = h.CreateChatMessage(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response handlers.ChatMessage
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestGetChatMessages(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []handlers.ChatMessage)
	}{
		{
			name: "Success - Default Pagination",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetChatMessages(gomock.Any(), db.GetChatMessagesParams{
						ListID: listID,
						Limit:  50,
						Offset: 0,
					}).
					Return([]db.GetChatMessagesRow{
						{
							ID:            uuid.New(),
							ListID:        listID,
							UserID:        userID,
							CharacterName: "TestCharacter",
							Message:       "Hello world",
							CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []handlers.ChatMessage) {
				require.Len(t, response, 1)
				require.Equal(t, "Hello world", response[0].Message)
				require.Equal(t, "TestCharacter", response[0].CharacterName)
			},
		},
		{
			name: "Success - Custom Pagination",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Set("limit", "10")
				c.QueryParams().Set("offset", "5")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetChatMessages(gomock.Any(), db.GetChatMessagesParams{
						ListID: listID,
						Limit:  10,
						Offset: 5,
					}).
					Return([]db.GetChatMessagesRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []handlers.ChatMessage) {
				require.Empty(t, response)
			},
		},
		{
			name: "Success - Since Parameter",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Set("since", "2023-01-01T00:00:00Z")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetChatMessagesByTimestamp(gomock.Any(), gomock.Any()).
					Return([]db.GetChatMessagesByTimestampRow{
						{
							ID:            uuid.New(),
							ListID:        listID,
							UserID:        userID,
							CharacterName: "TestCharacter",
							Message:       "Recent message",
							CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []handlers.ChatMessage) {
				require.Len(t, response, 1)
				require.Equal(t, "Recent message", response[0].Message)
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
			expectedError: "Invalid list ID",
		},
		{
			name: "User Not List Member",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "User is not a member of this list",
		},
		{
			name: "Invalid Since Timestamp",
			setupRequest: func(c echo.Context) {
				c.QueryParams().Set("since", "invalid-timestamp")
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid since timestamp",
		},
		{
			name: "Database Error - GetChatMessages",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					GetChatMessages(gomock.Any(), db.GetChatMessagesParams{
						ListID: listID,
						Limit:  50,
						Offset: 0,
					}).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get chat messages",
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
			req := httptest.NewRequest(http.MethodGet, "/api/lists/"+listID.String()+"/chat", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = validator.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/chat")
			c.SetParamNames("id")
			c.SetParamValues(listID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetChatMessages(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response []handlers.ChatMessage
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestDeleteChatMessage(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					DeleteChatMessage(gomock.Any(), db.DeleteChatMessageParams{
						ID:     messageID,
						UserID: userID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name: "Invalid Message ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("list-id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid message ID",
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Database Error - DeleteChatMessage",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, messageID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					DeleteChatMessage(gomock.Any(), db.DeleteChatMessageParams{
						ID:     messageID,
						UserID: userID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to delete chat message",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			listID := uuid.New()
			messageID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			url := fmt.Sprintf("/api/lists/%s/chat/%s", listID.String(), messageID.String())
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = validator.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/chat/:messageId")
			c.SetParamNames("id", "messageId")
			c.SetParamValues(listID.String(), messageID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, messageID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.DeleteChatMessage(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestGetChatNotifications(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response []handlers.ChatNotification)
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetChatNotificationsForUser(gomock.Any(), userID).
					Return([]db.GetChatNotificationsForUserRow{
						{
							ListID:            uuid.New(),
							ListName:          "Test List",
							LastMessageTime:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
							UnreadCount:       int64(3),
							LastCharacterName: "TestCharacter",
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []handlers.ChatNotification) {
				require.Len(t, response, 1)
				require.Equal(t, "Test List", response[0].ListName)
				require.Equal(t, int32(3), response[0].UnreadCount)
				require.Equal(t, "TestCharacter", response[0].LastCharacterName)
			},
		},
		{
			name: "No Notifications",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetChatNotificationsForUser(gomock.Any(), userID).
					Return([]db.GetChatNotificationsForUserRow{}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response []handlers.ChatNotification) {
				require.Empty(t, response)
			},
		},
		{
			name: "Invalid User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", "invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user ID format",
		},
		{
			name: "No User ID in Context",
			setupRequest: func(c echo.Context) {
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				// No mocks needed for this case
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid user authentication",
		},
		{
			name: "Database Error",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, userID uuid.UUID) {
				store.EXPECT().
					GetChatNotificationsForUser(gomock.Any(), userID).
					Return(nil, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get chat notifications",
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
			req := httptest.NewRequest(http.MethodGet, "/api/chat/notifications", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = validator.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/chat/notifications")
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.GetChatNotifications(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response []handlers.ChatNotification
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestMarkChatMessagesAsRead(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					MarkListMessagesAsRead(gomock.Any(), db.MarkListMessagesAsReadParams{
						UserID: userID,
						ListID: listID,
					}).
					Return(nil)
			},
			expectedCode: http.StatusNoContent,
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
			expectedError: "Invalid list ID",
		},
		{
			name: "User Not List Member",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, nil)
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "User is not a member of this list",
		},
		{
			name: "Database Error - IsUserListMember",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(false, errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to check list membership",
		},
		{
			name: "Database Error - MarkListMessagesAsRead",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, listID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					IsUserListMember(gomock.Any(), db.IsUserListMemberParams{
						ListID: listID,
						UserID: userID,
					}).
					Return(true, nil)

				store.EXPECT().
					MarkListMessagesAsRead(gomock.Any(), db.MarkListMessagesAsReadParams{
						UserID: userID,
						ListID: listID,
					}).
					Return(errors.New("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to mark messages as read",
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
			url := fmt.Sprintf("/api/lists/%s/chat/read", listID.String())
			req := httptest.NewRequest(http.MethodPost, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			e.Validator = validator.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/lists/:id/chat/read")
			c.SetParamNames("id")
			c.SetParamValues(listID.String())
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, listID, userID)

			// Execute handler
			h := handlers.NewListsHandler(store)
			err := h.MarkChatMessagesAsRead(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
