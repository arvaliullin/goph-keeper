package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestClient_Errors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("server error"))
		assert.NoError(t, err)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)

	_, err := client.Register("test", "pass")
	assert.Error(t, err)

	_, err = client.Login("test", "pass")
	assert.Error(t, err)

	_, err = client.CreateSecret(domain.SecretTypeLogin, []byte("d"), []byte("m"))
	assert.Error(t, err)

	_, err = client.ListSecrets()
	assert.Error(t, err)

	_, err = client.SyncSecrets(time.Now().Add(-time.Hour))
	assert.Error(t, err)

	_, err = client.GetSecret("1")
	assert.Error(t, err)

	_, err = client.UpdateSecret("1", []byte("d"), []byte("m"))
	assert.Error(t, err)

	err = client.DeleteSecret("1")
	assert.Error(t, err)

	_, err = client.DownloadBinary("1")
	assert.Error(t, err)

	err = client.DeleteBinary("1")
	assert.Error(t, err)
}

func TestClient_NetworkErrors(t *testing.T) {
	client := NewClient("http://127.0.0.1:0")

	_, err := client.Register("test", "pass")
	assert.Error(t, err)

	_, err = client.Login("test", "pass")
	assert.Error(t, err)

	_, err = client.CreateSecret(domain.SecretTypeLogin, []byte("d"), []byte("m"))
	assert.Error(t, err)

	_, err = client.ListSecrets()
	assert.Error(t, err)

	_, err = client.SyncSecrets(time.Now().Add(-time.Hour))
	assert.Error(t, err)

	_, err = client.GetSecret("1")
	assert.Error(t, err)

	_, err = client.UpdateSecret("1", []byte("d"), []byte("m"))
	assert.Error(t, err)

	err = client.DeleteSecret("1")
	assert.Error(t, err)

	err = client.DeleteBinary("1")
	assert.Error(t, err)
}
