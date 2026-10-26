package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergot/tibiacores/backend/pkg/apperror"
	"github.com/stretchr/testify/require"
)

func TestNewTibiaDataServiceUsesV5(t *testing.T) {
	service := NewTibiaDataService()

	require.Equal(t, "https://api.tibiadata.com/v5", service.baseURL)
}

func TestGetCharacter(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		body       string
		wantName   string
		wantType   apperror.ErrorType
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"character":{"character":{"name":"Test Character","world":"Antica","comment":"TIBIACORES-code"}}}`,
			wantName:   "Test Character",
		},
		{
			name:       "character not found",
			statusCode: http.StatusNotFound,
			wantType:   apperror.ErrorTypeNotFound,
		},
		{
			name:       "upstream failure",
			statusCode: http.StatusBadGateway,
			wantType:   apperror.ErrorTypeExternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestPath = r.URL.Path
				w.WriteHeader(tc.statusCode)
				_, err := fmt.Fprint(w, tc.body)
				require.NoError(t, err)
			}))
			defer server.Close()

			service := &TibiaDataService{baseURL: server.URL + "/v5", client: server.Client()}
			character, err := service.GetCharacter("TestCharacter")
			require.Equal(t, "/v5/character/TestCharacter", requestPath)
			if tc.wantType != "" {
				require.Error(t, err)
				var appErr *apperror.AppError
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tc.wantType, appErr.Type)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantName, character.Name)
			require.Equal(t, "Antica", character.World)
			require.Equal(t, "TIBIACORES-code", character.Comment)
		})
	}
}

func TestGetCharacterReturnsDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := fmt.Fprint(w, `invalid json`)
		require.NoError(t, err)
	}))
	defer server.Close()

	service := &TibiaDataService{baseURL: server.URL + "/v5", client: server.Client()}
	_, err := service.GetCharacter("TestCharacter")

	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, apperror.ErrorTypeExternal, appErr.Type)
}
