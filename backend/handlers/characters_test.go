package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergot/tibiacores/backend/db/mock"
	db "github.com/sergot/tibiacores/backend/db/sqlc"
)

func TestGetHighscores(t *testing.T) {
	// Test cases
	testCases := []struct {
		name          string
		queryParams   map[string]string
		setupMocks    func(store *mock.MockStore)
		expectedCode  int
		checkResponse func(t *testing.T, response HighscoreResponse)
	}{
		{
			name:        "Success with default pagination",
			queryParams: map[string]string{},
			setupMocks: func(store *mock.MockStore) {
				characters := []db.GetHighscoreCharactersRow{
					{
						ID:         uuid.New(),
						Name:       "Character1",
						World:      "Antica",
						CoreCount:  100,
						TotalCount: 43, // Total number of characters
					},
					{
						ID:         uuid.New(),
						Name:       "Character2",
						World:      "Secura",
						CoreCount:  90,
						TotalCount: 43,
					},
				}

				store.EXPECT().
					GetHighscoreCharacters(gomock.Any(), db.GetHighscoreCharactersParams{
						Limit:  20, // Default page size
						Offset: 0,  // First page
					}).
					Return(characters, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response HighscoreResponse) {
				require.Equal(t, 2, len(response.Characters))
				require.Equal(t, "Character1", response.Characters[0].Name)
				require.Equal(t, int64(100), response.Characters[0].CoreCount)
				require.Equal(t, "Character2", response.Characters[1].Name)
				require.Equal(t, int64(90), response.Characters[1].CoreCount)
				require.Equal(t, 3, response.Pagination.TotalPages) // 43 records / 20 per page = 3 pages
				require.Equal(t, 1, response.Pagination.CurrentPage)
				require.Equal(t, 43, response.Pagination.TotalRecords)
				require.Equal(t, 20, response.Pagination.PageSize)
			},
		},
		{
			name: "Success with custom page",
			queryParams: map[string]string{
				"page": "2",
			},
			setupMocks: func(store *mock.MockStore) {
				characters := []db.GetHighscoreCharactersRow{
					{
						ID:         uuid.New(),
						Name:       "Character3",
						World:      "Antica",
						CoreCount:  70,
						TotalCount: 43,
					},
					{
						ID:         uuid.New(),
						Name:       "Character4",
						World:      "Secura",
						CoreCount:  60,
						TotalCount: 43,
					},
				}

				store.EXPECT().
					GetHighscoreCharacters(gomock.Any(), db.GetHighscoreCharactersParams{
						Limit:  20, // Default page size
						Offset: 20, // Second page (offset = (page-1) * pageSize)
					}).
					Return(characters, nil)
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, response HighscoreResponse) {
				require.Equal(t, 2, len(response.Characters))
				require.Equal(t, "Character3", response.Characters[0].Name)
				require.Equal(t, int64(70), response.Characters[0].CoreCount)
				require.Equal(t, "Character4", response.Characters[1].Name)
				require.Equal(t, int64(60), response.Characters[1].CoreCount)
				require.Equal(t, 3, response.Pagination.TotalPages)
				require.Equal(t, 2, response.Pagination.CurrentPage)
				require.Equal(t, 43, response.Pagination.TotalRecords)
				require.Equal(t, 20, response.Pagination.PageSize)
			},
		},
		{
			name: "Invalid page parameter",
			queryParams: map[string]string{
				"page": "invalid",
			},
			setupMocks: func(store *mock.MockStore) {
				// No mock expectations - should fail validation before DB call
			},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response HighscoreResponse) {
				// Not checking response body as it will be an error message
			},
		},
		{
			name: "Page number too high",
			queryParams: map[string]string{
				"page": "51", // Max is 50
			},
			setupMocks: func(store *mock.MockStore) {
				// No mock expectations - should fail validation before DB call
			},
			expectedCode: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response HighscoreResponse) {
				// Not checking response body as it will be an error message
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock.NewMockStore(ctrl)
			handler := NewCharactersHandler(store)

			// Setup mock expectations
			tc.setupMocks(store)

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/highscores", nil)
			q := req.URL.Query()
			for key, value := range tc.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)

			// Execute handler
			err := handler.GetHighscores(c)

			// For error cases
			if tc.expectedCode >= 400 {
				require.Error(t, err)
				return
			}

			// For success cases
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)

			// Check response body
			var response HighscoreResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			tc.checkResponse(t, response)
		})
	}
}
