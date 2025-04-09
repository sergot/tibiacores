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
	"github.com/sergot/tibiacores/backend/middleware"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/sergot/tibiacores/backend/services"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockTibiaDataService struct {
	getCharacterFn         func(name string) (*services.TibiaCharacter, error)
	verifyCharacterClaimFn func(name, verificationCode string) (bool, error)
}

// Ensure mockTibiaDataService implements TibiaDataServiceInterface
var _ services.TibiaDataServiceInterface = (*mockTibiaDataService)(nil)

func (m *mockTibiaDataService) GetCharacter(name string) (*services.TibiaCharacter, error) {
	if m.getCharacterFn != nil {
		return m.getCharacterFn(name)
	}
	return nil, errors.New("GetCharacter not implemented")
}

func (m *mockTibiaDataService) VerifyCharacterClaim(name, verificationCode string) (bool, error) {
	if m.verifyCharacterClaimFn != nil {
		return m.verifyCharacterClaimFn(name, verificationCode)
	}
	return false, errors.New("VerifyCharacterClaim not implemented")
}

func TestStartClaim(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context, body *bytes.Buffer)
		setupMocks    func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response *handlers.StartClaimResponse)
	}{
		{
			name: "Success - Authenticated User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TestChar",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, userID uuid.UUID) {
				tibiaChar := &services.TibiaCharacter{
					Name:  "TestChar",
					World: "Antica",
				}
				tibiaData.getCharacterFn = func(name string) (*services.TibiaCharacter, error) {
					return tibiaChar, nil
				}

				charID := uuid.New()
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "TestChar").
					Return(db.Character{
						ID:   charID,
						Name: "TestChar",
					}, nil)

				store.EXPECT().
					GetCharacterClaim(gomock.Any(), db.GetCharacterClaimParams{
						CharacterID: charID,
						ClaimerID:   userID,
					}).
					Return(db.CharacterClaim{}, sql.ErrNoRows)

				store.EXPECT().
					CreateCharacterClaim(gomock.Any(), gomock.Any()).
					Return(db.CharacterClaim{
						ID:               uuid.New(),
						CharacterID:      charID,
						ClaimerID:        userID,
						VerificationCode: "TIBIACORES-1234",
						Status:           "pending",
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response *handlers.StartClaimResponse) {
				require.NotEmpty(t, response.ClaimID)
				require.NotEmpty(t, response.VerificationCode)
				require.Equal(t, "pending", response.Status)
			},
		},
		{
			name: "Success - Anonymous User",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "TestChar",
				})
				require.NoError(t, err)
				c.Set("user_id", nil)
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, userID uuid.UUID) {
				tibiaChar := &services.TibiaCharacter{
					Name:  "TestChar",
					World: "Antica",
				}
				tibiaData.getCharacterFn = func(name string) (*services.TibiaCharacter, error) {
					return tibiaChar, nil
				}

				charID := uuid.New()
				store.EXPECT().
					GetCharacterByName(gomock.Any(), "TestChar").
					Return(db.Character{
						ID:   charID,
						Name: "TestChar",
					}, nil)

				store.EXPECT().
					CreateAnonymousUser(gomock.Any(), gomock.Any()).
					Return(db.User{
						ID: userID,
					}, nil)

				store.EXPECT().
					GetCharacterClaim(gomock.Any(), db.GetCharacterClaimParams{
						CharacterID: charID,
						ClaimerID:   userID,
					}).
					Return(db.CharacterClaim{}, sql.ErrNoRows)

				store.EXPECT().
					CreateCharacterClaim(gomock.Any(), gomock.Any()).
					Return(db.CharacterClaim{
						ID:               uuid.New(),
						CharacterID:      charID,
						ClaimerID:        userID,
						VerificationCode: "TIBIACORES-1234",
						Status:           "pending",
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, response *handlers.StartClaimResponse) {
				require.NotEmpty(t, response.ClaimID)
				require.NotEmpty(t, response.VerificationCode)
				require.Equal(t, "pending", response.Status)
			},
		},
		{
			name: "Invalid Request Body",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				body.WriteString("{invalid json")
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Character Not Found in Tibia",
			setupRequest: func(c echo.Context, body *bytes.Buffer) {
				body.Reset()
				err := json.NewEncoder(body).Encode(map[string]interface{}{
					"character_name": "NonExistentChar",
				})
				require.NoError(t, err)
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, userID uuid.UUID) {
				tibiaData.getCharacterFn = func(name string) (*services.TibiaCharacter, error) {
					return nil, errors.New("character not found")
				}
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Character not found in Tibia",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tibiaData := &mockTibiaDataService{}
			userID := uuid.New()

			// Create HTTP request
			reqBody := bytes.NewBuffer([]byte(`{}`))
			req := httptest.NewRequest(http.MethodPost, "/api/claims", reqBody)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/claims")
			c.Set("user_id", userID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c, reqBody)
			}

			// Setup mock expectations
			tc.setupMocks(store, tibiaData, userID)

			// Create handler with mock store and tibia data service
			h := handlers.NewClaimsHandler(store)
			h.TibiaData = tibiaData

			// Execute handler
			err := h.StartClaim(c)

			// Check for expected error response
			if tc.expectedError != "" {
				// Use the ErrorHandler to process the error
				middleware.ErrorHandler(err, c)

				// Check if we received an error with the correct status code and message
				require.Equal(t, tc.expectedCode, rec.Code)

				var errorResponse map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errorResponse))
				require.Contains(t, errorResponse["message"].(string), tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response handlers.StartClaimResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, &response)
			}
		})
	}
}

func TestCheckClaim(t *testing.T) {
	testCases := []struct {
		name          string
		setupRequest  func(c echo.Context)
		setupMocks    func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, claimID uuid.UUID, userID uuid.UUID)
		expectedCode  int
		expectedError string
		checkResponse func(t *testing.T, response map[string]interface{})
	}{
		{
			name: "Success - Claim Verified",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, claimID uuid.UUID, userID uuid.UUID) {
				claim := db.GetClaimByIDRow{
					ID:               claimID,
					CharacterID:      uuid.New(),
					ClaimerID:        userID,
					CharacterName:    "TestChar",
					Status:           "pending",
					VerificationCode: "TIBIACORES-1234",
				}

				store.EXPECT().
					GetClaimByID(gomock.Any(), claimID).
					Return(claim, nil)

				tibiaData.verifyCharacterClaimFn = func(name, code string) (bool, error) {
					return true, nil
				}

				store.EXPECT().
					UpdateClaimStatus(gomock.Any(), db.UpdateClaimStatusParams{
						CharacterID: claim.CharacterID,
						ClaimerID:   claim.ClaimerID,
						Status:      "approved",
					}).
					Return(db.CharacterClaim{
						ID:               claim.ID,
						CharacterID:      claim.CharacterID,
						ClaimerID:        claim.ClaimerID,
						Status:           "approved",
						VerificationCode: claim.VerificationCode,
					}, nil)

				store.EXPECT().
					DeactivateCharacterListMemberships(gomock.Any(), claim.CharacterID).
					Return(nil)

				store.EXPECT().
					UpdateCharacterOwner(gomock.Any(), db.UpdateCharacterOwnerParams{
						ID:     claim.CharacterID,
						UserID: claim.ClaimerID,
					}).
					Return(db.Character{
						ID:     claim.CharacterID,
						UserID: claim.ClaimerID,
						Name:   claim.CharacterName,
					}, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				require.Equal(t, "approved", response["status"])
				require.NotNil(t, response["character"])
			},
		},
		{
			name: "Invalid Claim ID",
			setupRequest: func(c echo.Context) {
				c.SetParamValues("invalid-uuid")
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, claimID uuid.UUID, userID uuid.UUID) {
				// No mocks needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid claim ID",
		},
		{
			name: "Claim Not Found",
			setupRequest: func(c echo.Context) {
				// Default setup is fine
			},
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService, claimID uuid.UUID, userID uuid.UUID) {
				store.EXPECT().
					GetClaimByID(gomock.Any(), claimID).
					Return(db.GetClaimByIDRow{}, sql.ErrNoRows)
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "Claim not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tibiaData := &mockTibiaDataService{}
			claimID := uuid.New()
			userID := uuid.New()

			// Create HTTP request
			url := fmt.Sprintf("/api/claims/%s", claimID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Default context setup
			c.SetPath("/api/claims/:id")
			c.Set("user_id", userID.String())
			c.SetParamNames("id")
			c.SetParamValues(claimID.String())

			// Custom request setup if needed
			if tc.setupRequest != nil {
				tc.setupRequest(c)
			}

			// Setup mock expectations
			tc.setupMocks(store, tibiaData, claimID, userID)

			// Create handler with mock store and tibia data service
			h := handlers.NewClaimsHandler(store)
			h.TibiaData = tibiaData

			// Execute handler
			err := h.CheckClaim(c)

			// Check for expected error response
			if tc.expectedError != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tc.expectedCode, appErr.StatusCode)
				require.Contains(t, appErr.Message, tc.expectedError)
				return
			}

			// Check successful response
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			if tc.checkResponse != nil {
				var response map[string]interface{}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
				tc.checkResponse(t, response)
			}
		})
	}
}

func TestProcessPendingClaims(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(store *mockdb.MockStore, tibiaData *mockTibiaDataService)
	}{
		{
			name: "Success - Process Multiple Claims",
			setupMocks: func(store *mockdb.MockStore, tibiaData *mockTibiaDataService) {
				// Create two test claims
				expiredClaim := db.GetPendingClaimsToCheckRow{
					ID:               uuid.New(),
					CharacterID:      uuid.New(),
					ClaimerID:        uuid.New(),
					CharacterName:    "ExpiredCharacter",
					VerificationCode: "TIBIACORES-1234",
					Status:           "pending",
					CreatedAt:        pgtype.Timestamptz{Time: time.Now().Add(-25 * time.Hour), Valid: true},
				}
				validClaim := db.GetPendingClaimsToCheckRow{
					ID:               uuid.New(),
					CharacterID:      uuid.New(),
					ClaimerID:        uuid.New(),
					CharacterName:    "ValidCharacter",
					VerificationCode: "TIBIACORES-5678",
					Status:           "pending",
					CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
				}

				store.EXPECT().
					GetPendingClaimsToCheck(gomock.Any()).
					Return([]db.GetPendingClaimsToCheckRow{expiredClaim, validClaim}, nil)

				// For expired claim
				store.EXPECT().
					GetCharacter(gomock.Any(), expiredClaim.CharacterID).
					Return(db.Character{
						ID:   expiredClaim.CharacterID,
						Name: expiredClaim.CharacterName,
					}, nil)

				tibiaData.verifyCharacterClaimFn = func(name, code string) (bool, error) {
					return false, nil
				}

				// Handle both the initial "pending" update and final "rejected" update for expired claim
				store.EXPECT().
					UpdateClaimStatus(gomock.Any(), gomock.Any()).
					Return(db.CharacterClaim{
						ID:               expiredClaim.ID,
						CharacterID:      expiredClaim.CharacterID,
						ClaimerID:        expiredClaim.ClaimerID,
						Status:           "rejected",
						VerificationCode: expiredClaim.VerificationCode,
					}, nil).
					AnyTimes()

				// For valid claim
				store.EXPECT().
					GetCharacter(gomock.Any(), validClaim.CharacterID).
					Return(db.Character{
						ID:   validClaim.CharacterID,
						Name: validClaim.CharacterName,
					}, nil)

				// Handle both the initial "pending" update and final "approved" update for valid claim
				store.EXPECT().
					UpdateClaimStatus(gomock.Any(), gomock.Any()).
					Return(db.CharacterClaim{
						ID:               validClaim.ID,
						CharacterID:      validClaim.CharacterID,
						ClaimerID:        validClaim.ClaimerID,
						Status:           "approved",
						VerificationCode: validClaim.VerificationCode,
					}, nil).
					AnyTimes()

				store.EXPECT().
					DeactivateCharacterListMemberships(gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				store.EXPECT().
					UpdateCharacterOwner(gomock.Any(), gomock.Any()).
					Return(db.Character{}, nil).
					AnyTimes()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tibiaData := &mockTibiaDataService{}

			tc.setupMocks(store, tibiaData)

			h := handlers.NewClaimsHandler(store)
			h.TibiaData = tibiaData

			err := h.ProcessPendingClaims()
			require.NoError(t, err)
		})
	}
}
