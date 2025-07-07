package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	mockdb "github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewsletterHandler_Subscribe(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func() *http.Request
		setupMocks    func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, body string)
	}{
		{
			name: "Success - New subscription",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// First check if subscriber exists - return error (not found)
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{}, fmt.Errorf("not found"))

				// Subscribe to EmailOctopus
				service.EXPECT().SubscribeToNewsletter(gomock.Any(), "test@example.com").
					Return("contact-123", nil)

				// Create subscriber in database
				contactID := pgtype.Text{String: "contact-123", Valid: true}
				store.EXPECT().CreateNewsletterSubscriber(gomock.Any(), db.CreateNewsletterSubscriberParams{
					Email:                 "test@example.com",
					EmailoctopusContactID: contactID,
				}).Return(db.NewsletterSubscriber{
					Email:                 "test@example.com",
					EmailoctopusContactID: contactID,
					Confirmed:             false,
				}, nil)

				// Confirm subscription
				store.EXPECT().ConfirmNewsletterSubscription(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:     "test@example.com",
						Confirmed: true,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "subscribed", response["status"])
				assert.Contains(t, response["message"], "Successfully subscribed")
			},
		},
		{
			name: "Success - Already subscribed",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Subscriber already exists and is confirmed
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "test@example.com",
						Confirmed:      true,
						UnsubscribedAt: pgtype.Timestamptz{Valid: false},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "already_subscribed", response["status"])
				assert.Contains(t, response["message"], "Already subscribed")
			},
		},
		{
			name: "Success - Resubscribe after unsubscribing",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Subscriber exists but was unsubscribed
				unsubscribedTime := pgtype.Timestamptz{Time: time.Now(), Valid: true}
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "test@example.com",
						Confirmed:      true,
						UnsubscribedAt: unsubscribedTime,
					}, nil)

				// Resubscribe to EmailOctopus
				service.EXPECT().SubscribeToNewsletter(gomock.Any(), "test@example.com").
					Return("contact-123", nil)

				// Confirm subscription (clears unsubscribed_at)
				store.EXPECT().ConfirmNewsletterSubscription(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:     "test@example.com",
						Confirmed: true,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "resubscribed", response["status"])
				assert.Contains(t, response["message"], "resubscribed")
			},
		},
		{
			name: "Error - Invalid JSON",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// No mocks needed for this test
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Error - Empty email",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// No mocks needed for this test
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email is required",
		},
		{
			name: "Error - EmailOctopus service failure",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Check if subscriber exists - return error (not found)
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{}, fmt.Errorf("not found"))

				// EmailOctopus service fails
				service.EXPECT().SubscribeToNewsletter(gomock.Any(), "test@example.com").
					Return("", fmt.Errorf("service unavailable"))
			},
			expectedCode:  http.StatusBadGateway,
			expectedError: "Failed to subscribe to newsletter",
		},
		{
			name: "Error - Database failure when creating subscriber",
			setupRequest: func() *http.Request {
				reqBody := SubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Check if subscriber exists - return error (not found)
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{}, fmt.Errorf("not found"))

				// Subscribe to EmailOctopus
				service.EXPECT().SubscribeToNewsletter(gomock.Any(), "test@example.com").
					Return("contact-123", nil)

				// Database fails to create subscriber
				contactID := pgtype.Text{String: "contact-123", Valid: true}
				store.EXPECT().CreateNewsletterSubscriber(gomock.Any(), db.CreateNewsletterSubscriberParams{
					Email:                 "test@example.com",
					EmailoctopusContactID: contactID,
				}).Return(db.NewsletterSubscriber{}, fmt.Errorf("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to store subscription",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			store := mockdb.NewMockStore(ctrl)
			service := mock.NewMockNewsletterServiceInterface(ctrl)

			// Setup mocks
			tc.setupMocks(store, service)

			// Create handler
			handler := NewNewsletterHandler(store, service)

			// Setup HTTP request/response
			req := tc.setupRequest()
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Execute test
			err := handler.Subscribe(c)

			if tc.expectedError != "" {
				// Check error case
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				// Check success case
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)
				if tc.checkResponse != nil {
					tc.checkResponse(t, rec.Body.String())
				}
			}
		})
	}
}

func TestNewsletterHandler_Unsubscribe(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func() *http.Request
		setupMocks    func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, body string)
	}{
		{
			name: "Success - Unsubscribe with contact ID",
			setupRequest: func() *http.Request {
				reqBody := UnsubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/unsubscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Get subscriber by email
				contactID := pgtype.Text{String: "contact-123", Valid: true}
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:                 "test@example.com",
						EmailoctopusContactID: contactID,
						Confirmed:             true,
					}, nil)

				// Unsubscribe from EmailOctopus
				service.EXPECT().UnsubscribeFromNewsletterByID(gomock.Any(), "contact-123").
					Return(nil)

				// Update database
				store.EXPECT().UnsubscribeFromNewsletter(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "test@example.com",
						UnsubscribedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "unsubscribed", response["status"])
				assert.Contains(t, response["message"], "Successfully unsubscribed")
			},
		},
		{
			name: "Success - Unsubscribe without contact ID",
			setupRequest: func() *http.Request {
				reqBody := UnsubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/unsubscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Get subscriber by email (no contact ID)
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:                 "test@example.com",
						EmailoctopusContactID: pgtype.Text{Valid: false},
						Confirmed:             true,
					}, nil)

				// Update database (EmailOctopus unsubscribe skipped due to no contact ID)
				store.EXPECT().UnsubscribeFromNewsletter(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "test@example.com",
						UnsubscribedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, "unsubscribed", response["status"])
			},
		},
		{
			name: "Error - Invalid JSON",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/newsletter/unsubscribe", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// No mocks needed for this test
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Error - Email not found",
			setupRequest: func() *http.Request {
				reqBody := UnsubscribeRequest{
					Email: "notfound@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/unsubscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Email not found in database
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "notfound@example.com").
					Return(db.NewsletterSubscriber{}, fmt.Errorf("not found"))
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email not found in newsletter",
		},
		{
			name: "Error - EmailOctopus service failure",
			setupRequest: func() *http.Request {
				reqBody := UnsubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/newsletter/unsubscribe", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupMocks: func(store *mockdb.MockStore, service *mock.MockNewsletterServiceInterface) {
				// Get subscriber by email
				contactID := pgtype.Text{String: "contact-123", Valid: true}
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:                 "test@example.com",
						EmailoctopusContactID: contactID,
						Confirmed:             true,
					}, nil)

				// EmailOctopus service fails
				service.EXPECT().UnsubscribeFromNewsletterByID(gomock.Any(), "contact-123").
					Return(fmt.Errorf("service unavailable"))
			},
			expectedCode:  http.StatusBadGateway,
			expectedError: "Failed to unsubscribe from newsletter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			store := mockdb.NewMockStore(ctrl)
			service := mock.NewMockNewsletterServiceInterface(ctrl)

			// Setup mocks
			tc.setupMocks(store, service)

			// Create handler
			handler := NewNewsletterHandler(store, service)

			// Setup HTTP request/response
			req := tc.setupRequest()
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Execute test
			err := handler.Unsubscribe(c)

			if tc.expectedError != "" {
				// Check error case
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				// Check success case
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)
				if tc.checkResponse != nil {
					tc.checkResponse(t, rec.Body.String())
				}
			}
		})
	}
}

func TestNewsletterHandler_CheckSubscriptionStatus(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func() *http.Request
		setupMocks    func(store *mockdb.MockStore)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, body string)
	}{
		{
			name: "Success - Subscribed user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/newsletter/status?email=test@example.com", nil)
				return req
			},
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "test@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "test@example.com",
						Confirmed:      true,
						UnsubscribedAt: pgtype.Timestamptz{Valid: false},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, true, response["subscribed"])
			},
		},
		{
			name: "Success - Not subscribed user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/newsletter/status?email=notfound@example.com", nil)
				return req
			},
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "notfound@example.com").
					Return(db.NewsletterSubscriber{}, fmt.Errorf("not found"))
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, false, response["subscribed"])
			},
		},
		{
			name: "Success - Unsubscribed user",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/newsletter/status?email=unsubscribed@example.com", nil)
				return req
			},
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().GetNewsletterSubscriberByEmail(gomock.Any(), "unsubscribed@example.com").
					Return(db.NewsletterSubscriber{
						Email:          "unsubscribed@example.com",
						Confirmed:      true,
						UnsubscribedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, false, response["subscribed"])
			},
		},
		{
			name: "Error - Missing email parameter",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/newsletter/status", nil)
				return req
			},
			setupMocks: func(store *mockdb.MockStore) {
				// No mocks needed for this test
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Email is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			store := mockdb.NewMockStore(ctrl)

			// Setup mocks
			tc.setupMocks(store)

			// Create handler (newsletter service not needed for this endpoint)
			handler := NewNewsletterHandler(store, nil)

			// Setup HTTP request/response
			req := tc.setupRequest()
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Execute test
			err := handler.CheckSubscriptionStatus(c)

			if tc.expectedError != "" {
				// Check error case
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				// Check success case
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)
				if tc.checkResponse != nil {
					tc.checkResponse(t, rec.Body.String())
				}
			}
		})
	}
}

func TestNewsletterHandler_GetStats(t *testing.T) {
	testCases := []struct {
		name          string
		setupMocks    func(store *mockdb.MockStore)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, body string)
	}{
		{
			name: "Success - Get stats",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().GetNewsletterSubscriberStats(gomock.Any()).
					Return(db.GetNewsletterSubscriberStatsRow{
						TotalSubscribers:    100,
						ActiveSubscribers:   85,
						PendingConfirmation: 5,
						Unsubscribed:        10,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var response map[string]interface{}
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)
				assert.Equal(t, float64(100), response["total_subscribers"])
				assert.Equal(t, float64(85), response["active_subscribers"])
				assert.Equal(t, float64(5), response["pending_confirmation"])
				assert.Equal(t, float64(10), response["unsubscribed"])
			},
		},
		{
			name: "Error - Database failure",
			setupMocks: func(store *mockdb.MockStore) {
				store.EXPECT().GetNewsletterSubscriberStats(gomock.Any()).
					Return(db.GetNewsletterSubscriberStatsRow{}, fmt.Errorf("database error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get newsletter stats",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			store := mockdb.NewMockStore(ctrl)

			// Setup mocks
			tc.setupMocks(store)

			// Create handler (newsletter service not needed for this endpoint)
			handler := NewNewsletterHandler(store, nil)

			// Setup HTTP request/response
			req := httptest.NewRequest(http.MethodGet, "/newsletter/stats", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Execute test
			err := handler.GetStats(c)

			if tc.expectedError != "" {
				// Check error case
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				// Check success case
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)
				if tc.checkResponse != nil {
					tc.checkResponse(t, rec.Body.String())
				}
			}
		})
	}
}
