package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/goph-keeper/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authService := mocks.NewMockAuthService(ctrl)
	secretService := mocks.NewMockSecretService(ctrl)
	binaryService := mocks.NewMockBinaryService(ctrl)

	r := NewRouter(authService, secretService, binaryService, "secret", 1024)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/swagger/index.html")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp2, err := http.Post(ts.URL+"/api/v1/register", "application/json", nil)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusNotFound, resp2.StatusCode)
}
