package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func (m *mockOAuthProvider) ValidateState(state string) bool {
	return m.validateState
}

func (m *mockOAuthProvider) ExchangeCode(provider, code string) (*auth.OAuthUserInfo, error) {
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
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)

				var resp map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.True(t, resp["has_email"].(bool))
			},
		},
		{
			name: "Database Error - GetUserByEmail",
			setupMock: func(store *mock.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), pgtype.Text{String: "test@example.com", Valid: true}).
					Return(db.User{}, sql.ErrConnDone)
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(db.User{}, sql.ErrConnDone)
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
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)

				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, "Failed to create user", appErr.Message)
				assert.Equal(t, sql.ErrConnDone, appErr.Unwrap())
			},
		},
		{
			name:      "Missing Email from Provider",
			setupMock: func(store *mock.MockStore) {},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: true,
					userInfo: &auth.OAuthUserInfo{
						Email: "",
					},
					err: apperror.ValidationError("Email is required", nil),
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "test",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)

				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, "Failed to authenticate with provider", appErr.Message)
			},
		},
		{
			name:      "Invalid Provider",
			setupMock: func(store *mock.MockStore) {},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: true,
					userInfo: &auth.OAuthUserInfo{
						Email: "test@example.com",
					},
					err: apperror.ValidationError("Invalid OAuth provider", nil),
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "test",
			},
			pathParams: map[string]string{
				"provider": "invalid",
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)

				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, "Failed to authenticate with provider", appErr.Message)
			},
		},
		{
			name:      "Invalid State",
			setupMock: func(store *mock.MockStore) {},
			setupProvider: func() *mockOAuthProvider {
				return &mockOAuthProvider{
					validateState: false,
					userInfo: &auth.OAuthUserInfo{
						Email: "test@example.com",
					},
				}
			},
			queryParams: map[string]string{
				"code":  "test",
				"state": "invalid",
			},
			pathParams: map[string]string{
				"provider": "google",
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder, err error) {
				require.Error(t, err)

				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, "Invalid OAuth state", appErr.Message)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock.NewMockStore(ctrl)
			tc.setupMock(mockStore)

			handler := NewOAuthHandler(mockStore)
			handler.SetOAuthProvider(tc.setupProvider())

			rec := httptest.NewRecorder()

			// Build query string
			query := ""
			for k, v := range tc.queryParams {
				if query != "" {
					query += "&"
				}
				query += k + "=" + v
			}

			req := httptest.NewRequest(http.MethodGet, "/oauth/callback?"+query, nil)
			c := e.NewContext(req, rec)
			c.SetPath("/oauth/:provider/callback")
			c.SetParamNames("provider")
			c.SetParamValues(tc.pathParams["provider"])

			err := handler.Callback(c)
			tc.checkResponse(t, rec, err)
		})
	}
}
