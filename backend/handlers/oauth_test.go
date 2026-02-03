package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/sergot/tibiacores/backend/auth"
	"github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type mockOAuthProvider struct {
	auth.OAuthProvider
	validateState bool
	userInfo      *auth.OAuthUserInfo
	err           error
}

func (m *mockOAuthProvider) ValidateState(cookieState, queryState string) bool {
	return m.validateState
}

func (m *mockOAuthProvider) ExchangeCode(ctx context.Context, provider, code string) (*auth.OAuthUserInfo, error) {
	return m.userInfo, m.err
}

func TestOAuthHandler_Callback(t *testing.T) {
	e := echo.New()

	testCases := []struct {
		name          string
		setupMock     func(*mock.MockStore)
		setupProvider func() *mockOAuthProvider
		queryParams   map[string]string
		pathParams    map[string]string
		cookieState   string
		checkResponse func(*testing.T, *httptest.ResponseRecorder, error)
	}{
		{
			name: "Success - New User",
			setupMock: func(store *mock.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), pgtype.Text{String: "test@example.com", Valid: true}).
					Return(db.User{}, sql.ErrNoRows)
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{
						ID:    uuid.New(),
						Email: pgtype.Text{String: "test@example.com", Valid: true},
					}, nil)
			},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: true,
					userInfo: &auth.OAuthUserInfo{
						Email: "test@example.com",
					},
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "test",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			cookieState: "test",
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)

				var resp map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.True(t, resp["has_email"].(bool))
			},
		},
		{
			name: "Success - Existing User",
			setupMock: func(store *mock.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), pgtype.Text{String: "test@example.com", Valid: true}).
					Return(db.User{
						ID:    uuid.New(),
						Email: pgtype.Text{String: "test@example.com", Valid: true},
					}, nil)
			},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: true,
					userInfo: &auth.OAuthUserInfo{
						Email: "test@example.com",
					},
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "test",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			cookieState: "test",
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)

				var resp map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.True(t, resp["has_email"].(bool))
			},
		},
		{
			name:      "Invalid State",
			setupMock: func(store *mock.MockStore) {},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: false,
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "invalid",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			cookieState: "test",
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)
				assert.IsType(t, &apperror.AppError{}, err)
				appErr := err.(*apperror.AppError)
				assert.Equal(t, apperror.ErrorTypeValidation, appErr.Type)
			},
		},
		{
			name: "Account Type Mismatch",
			setupMock: func(store *mock.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), pgtype.Text{String: "test@example.com", Valid: true}).
					Return(db.User{
						ID:       uuid.New(),
						Email:    pgtype.Text{String: "test@example.com", Valid: true},
						Password: pgtype.Text{String: "hashed_password", Valid: true},
					}, nil)
			},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: true,
					userInfo: &auth.OAuthUserInfo{
						Email: "test@example.com",
					},
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "test",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			cookieState: "test",
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)
				assert.IsType(t, &apperror.AppError{}, err)
				appErr := err.(*apperror.AppError)
				assert.Equal(t, apperror.ErrorTypeValidation, appErr.Type)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock.NewMockStore(ctrl)
			tc.setupMock(store)

			handler := NewOAuthHandler(store)
			handler.SetOAuthProvider(tc.setupProvider())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.cookieState != "" {
				req.AddCookie(&http.Cookie{
					Name:     "oauth_state",
					Value:    tc.cookieState,
					Expires:  time.Now().Add(1 * time.Hour),
					HttpOnly: true,
				})
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			q := c.Request().URL.Query()
			for k, v := range tc.queryParams {
				q.Add(k, v)
			}
			c.Request().URL.RawQuery = q.Encode()

			for k, v := range tc.pathParams {
				c.SetParamNames(k)
				c.SetParamValues(v)
			}

			err := handler.Callback(c)
			tc.checkResponse(t, rec, err)
		})
	}
}
