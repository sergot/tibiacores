package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/handlers"
	"github.com/sergot/tibiacores/backend/pkg/validator"
	"github.com/sergot/tibiacores/backend/services/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewsletterHandler_Subscribe(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func() *bytes.Buffer
		setupMocks    func(newsletterService *mock.MockNewsletterServiceInterface)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response handlers.NewsletterSubscribeResponse)
	}{
		{
			name: "success",
			setupRequest: func() *bytes.Buffer {
				reqBody := handlers.NewsletterSubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				return bytes.NewBuffer(jsonBody)
			},
			setupMocks: func(newsletterService *mock.MockNewsletterServiceInterface) {
				newsletterService.EXPECT().Subscribe(gomock.Any(), "test@example.com").Return(nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response handlers.NewsletterSubscribeResponse) {
				assert.Equal(t, "Successfully subscribed to newsletter", response.Message)
			},
		},
		{
			name: "invalid email",
			setupRequest: func() *bytes.Buffer {
				reqBody := handlers.NewsletterSubscribeRequest{
					Email: "invalid-email",
				}
				jsonBody, _ := json.Marshal(reqBody)
				return bytes.NewBuffer(jsonBody)
			},
			setupMocks: func(newsletterService *mock.MockNewsletterServiceInterface) {
				// No expectations - validation should fail before service call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid email address",
		},
		{
			name: "empty email",
			setupRequest: func() *bytes.Buffer {
				reqBody := handlers.NewsletterSubscribeRequest{
					Email: "",
				}
				jsonBody, _ := json.Marshal(reqBody)
				return bytes.NewBuffer(jsonBody)
			},
			setupMocks: func(newsletterService *mock.MockNewsletterServiceInterface) {
				// No expectations - validation should fail before service call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid email address",
		},
		{
			name: "invalid JSON",
			setupRequest: func() *bytes.Buffer {
				return bytes.NewBuffer([]byte("invalid json"))
			},
			setupMocks: func(newsletterService *mock.MockNewsletterServiceInterface) {
				// No expectations - JSON parsing should fail before service call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "service error",
			setupRequest: func() *bytes.Buffer {
				reqBody := handlers.NewsletterSubscribeRequest{
					Email: "test@example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				return bytes.NewBuffer(jsonBody)
			},
			setupMocks: func(newsletterService *mock.MockNewsletterServiceInterface) {
				newsletterService.EXPECT().Subscribe(gomock.Any(), "test@example.com").Return(assert.AnError)
			},
			expectedCode:  http.StatusBadGateway,
			expectedError: "Failed to subscribe to newsletter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock newsletter service
			newsletterService := mock.NewMockNewsletterServiceInterface(ctrl)
			tc.setupMocks(newsletterService)

			// Create handler
			handler := handlers.NewNewsletterHandler(newsletterService)

			// Setup HTTP request/response
			e := echo.New()
			e.Validator = validator.New()
			req := httptest.NewRequest(http.MethodPost, "/newsletter/subscribe", tc.setupRequest())
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute test
			err := handler.Subscribe(c)

			// Verify results
			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCode, rec.Code)

				if tc.checkResponse != nil {
					var response handlers.NewsletterSubscribeResponse
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					tc.checkResponse(t, response)
				}
			}
		})
	}
}
