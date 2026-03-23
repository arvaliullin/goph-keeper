package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestClient_Register(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/register", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(authResponse{Token: "token123"})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	token, err := client.Register("test", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "token123", token)
}

func TestClient_Login(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/login", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(authResponse{Token: "token321"})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	token, err := client.Login("test", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "token321", token)
}

func TestClient_CreateSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(domain.Secret{ID: "sec123"})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	client.SetToken("test_token")
	secret, err := client.CreateSecret(domain.SecretTypeLogin, []byte("data"), []byte("meta"))
	assert.NoError(t, err)
	assert.Equal(t, "sec123", secret.ID)
}

func TestClient_ListSecrets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode([]domain.Secret{{ID: "sec1"}, {ID: "sec2"}})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	secrets, err := client.ListSecrets()
	assert.NoError(t, err)
	assert.Len(t, secrets, 2)
}

func TestClient_SyncSecrets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("updated_after"))
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode([]domain.Secret{{ID: "sec1"}})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	secrets, err := client.SyncSecrets(time.Now().Add(-time.Hour))
	assert.NoError(t, err)
	assert.Len(t, secrets, 1)
}

func TestClient_GetSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets/1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(domain.Secret{ID: "1"})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	secret, err := client.GetSecret("1")
	assert.NoError(t, err)
	assert.Equal(t, "1", secret.ID)
}

func TestClient_UpdateSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/api/v1/secrets/sec123", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(domain.Secret{ID: "sec123"})
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	client.SetToken("test_token")
	secret, err := client.UpdateSecret("sec123", []byte("new_data"), []byte("new_meta"))
	assert.NoError(t, err)
	assert.Equal(t, "sec123", secret.ID)
}

func TestClient_DeleteSecret(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/secrets/1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	err := client.DeleteSecret("1")
	assert.NoError(t, err)
}

func TestClient_DownloadBinary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/binary/1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("binary_data"))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	data, err := client.DownloadBinary("1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("binary_data"), data)
}

func TestClient_DeleteBinary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/binary/1", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	err := client.DeleteBinary("1")
	assert.NoError(t, err)
}
