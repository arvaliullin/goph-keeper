//go:build integration

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/client/api"
	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/services"
	"github.com/arvaliullin/goph-keeper/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type integrationUserRepo struct {
	mu      sync.Mutex
	nextID  int64
	byLogin map[string]*domain.User
}

func newIntegrationUserRepo() *integrationUserRepo {
	return &integrationUserRepo{
		nextID:  1,
		byLogin: make(map[string]*domain.User),
	}
}

func (r *integrationUserRepo) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byLogin[user.Login]; exists {
		return domain.ErrUserAlreadyExists
	}

	user.ID = r.nextID
	user.CreatedAt = time.Now()
	r.nextID++
	r.byLogin[user.Login] = &domain.User{
		ID:           user.ID,
		Login:        user.Login,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}

	return nil
}

func (r *integrationUserRepo) GetByLogin(_ context.Context, login string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.byLogin[login]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	copyUser := *user
	return &copyUser, nil
}

type integrationSecretRepo struct {
	mu      sync.Mutex
	secrets map[string]*domain.Secret
}

func newIntegrationSecretRepo() *integrationSecretRepo {
	return &integrationSecretRepo{
		secrets: make(map[string]*domain.Secret),
	}
}

func (r *integrationSecretRepo) Create(_ context.Context, secret *domain.Secret) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if secret.CreatedAt.IsZero() {
		secret.CreatedAt = now
	}
	secret.UpdatedAt = now

	copySecret := *secret
	r.secrets[secret.ID] = &copySecret
	return nil
}

func (r *integrationSecretRepo) Update(_ context.Context, secret *domain.Secret) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.secrets[secret.ID]
	if !ok || existing.UserID != secret.UserID || existing.DeletedAt != nil {
		return domain.ErrSecretNotFound
	}

	existing.Data = slices.Clone(secret.Data)
	existing.Metadata = slices.Clone(secret.Metadata)
	existing.UpdatedAt = time.Now()
	return nil
}

func (r *integrationSecretRepo) Delete(_ context.Context, id string, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.secrets[id]
	if !ok || existing.UserID != userID || existing.DeletedAt != nil {
		return domain.ErrSecretNotFound
	}

	deletedAt := time.Now()
	existing.DeletedAt = &deletedAt
	existing.UpdatedAt = deletedAt
	return nil
}

func (r *integrationSecretRepo) GetByID(_ context.Context, id string, userID int64) (*domain.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.secrets[id]
	if !ok || existing.UserID != userID || existing.DeletedAt != nil {
		return nil, domain.ErrSecretNotFound
	}

	copySecret := *existing
	copySecret.Data = slices.Clone(existing.Data)
	copySecret.Metadata = slices.Clone(existing.Metadata)
	return &copySecret, nil
}

func (r *integrationSecretRepo) ListByUserID(_ context.Context, userID int64) ([]*domain.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var secrets []*domain.Secret
	for _, secret := range r.secrets {
		if secret.UserID != userID || secret.DeletedAt != nil {
			continue
		}
		copySecret := *secret
		copySecret.Data = slices.Clone(secret.Data)
		copySecret.Metadata = slices.Clone(secret.Metadata)
		secrets = append(secrets, &copySecret)
	}
	return secrets, nil
}

func (r *integrationSecretRepo) SyncByUserID(_ context.Context, userID int64, updatedAfter time.Time) ([]*domain.Secret, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var secrets []*domain.Secret
	for _, secret := range r.secrets {
		if secret.UserID != userID || !secret.UpdatedAt.After(updatedAfter) {
			continue
		}
		copySecret := *secret
		copySecret.Data = slices.Clone(secret.Data)
		copySecret.Metadata = slices.Clone(secret.Metadata)
		secrets = append(secrets, &copySecret)
	}
	return secrets, nil
}

type integrationBinaryRepo struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newIntegrationBinaryRepo() *integrationBinaryRepo {
	return &integrationBinaryRepo{
		objects: make(map[string][]byte),
	}
}

func (r *integrationBinaryRepo) Upload(_ context.Context, objectName string, reader io.Reader, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	r.objects[objectName] = data
	return nil
}

func (r *integrationBinaryRepo) Download(_ context.Context, objectName string) (io.ReadCloser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, ok := r.objects[objectName]
	if !ok {
		return nil, domain.ErrSecretNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (r *integrationBinaryRepo) Delete(_ context.Context, objectName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.objects, objectName)
	return nil
}

type integrationEnv struct {
	server  *httptest.Server
	client  *api.Client
	baseURL string
}

func setupIntegration(t *testing.T) *integrationEnv {
	t.Helper()

	userRepo := newIntegrationUserRepo()
	secretRepo := newIntegrationSecretRepo()
	binaryRepo := newIntegrationBinaryRepo()
	cryptoService, err := crypto.New("12345678901234567890123456789012")
	require.NoError(t, err)

	authService := services.NewAuthService(userRepo, "integration-secret")
	secretService := services.NewSecretService(secretRepo, binaryRepo, cryptoService)
	binaryService := services.NewBinaryService(binaryRepo, secretRepo, cryptoService)
	router := NewRouter(authService, secretService, binaryService, "integration-secret", 1024*1024)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &integrationEnv{
		server:  server,
		client:  api.NewClient(server.URL),
		baseURL: server.URL,
	}
}

func registerAndAuth(t *testing.T, env *integrationEnv, login, password string) {
	t.Helper()
	token, err := env.client.Register(login, password)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	env.client.SetToken(token)
}

func TestIntegration_RegisterAndLogin(t *testing.T) {
	env := setupIntegration(t)

	token, err := env.client.Register("user1", "password1")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	token2, err := env.client.Login("user1", "password1")
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	_, err = env.client.Register("user1", "other")
	assert.Error(t, err, "duplicate register should fail")

	_, err = env.client.Login("user1", "wrong")
	assert.Error(t, err, "wrong password should fail")

	_, err = env.client.Login("nonexistent", "pass")
	assert.Error(t, err, "nonexistent user should fail")
}

func TestIntegration_AddSecret_Login(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-login", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeLogin, []byte(`{"login":"admin","password":"secret"}`), []byte(`{"site":"example.com"}`))
	require.NoError(t, err)
	assert.NotEmpty(t, secret.ID)
	assert.Equal(t, domain.SecretTypeLogin, secret.Type)

	got, err := env.client.GetSecret(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, `{"login":"admin","password":"secret"}`, string(got.Data))
	assert.Equal(t, `{"site":"example.com"}`, string(got.Metadata))
}

func TestIntegration_AddSecret_Text(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-text", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("my secret note"), []byte("personal"))
	require.NoError(t, err)
	assert.Equal(t, domain.SecretTypeText, secret.Type)

	got, err := env.client.GetSecret(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "my secret note", string(got.Data))
}

func TestIntegration_AddSecret_Card(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-card", "pass")

	cardData := []byte(`{"number":"4111111111111111","exp":"12/28","cvv":"123"}`)
	secret, err := env.client.CreateSecret(domain.SecretTypeCard, cardData, []byte("visa"))
	require.NoError(t, err)
	assert.Equal(t, domain.SecretTypeCard, secret.Type)

	got, err := env.client.GetSecret(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, string(cardData), string(got.Data))
	assert.Equal(t, "visa", string(got.Metadata))
}

func TestIntegration_AddBinary(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-binary", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeBinary, nil, []byte("photo.jpg"))
	require.NoError(t, err)

	binaryPayload := []byte("fake-binary-content-here")
	err = env.client.UploadBinary(secret.ID, bytes.NewReader(binaryPayload), int64(len(binaryPayload)))
	require.NoError(t, err)

	downloaded, err := env.client.DownloadBinary(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, binaryPayload, downloaded)
}

func TestIntegration_UpdateSecret(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-update", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("original"), []byte("v1"))
	require.NoError(t, err)

	updated, err := env.client.UpdateSecret(secret.ID, []byte("modified"), []byte("v2"))
	require.NoError(t, err)
	assert.Equal(t, secret.ID, updated.ID)

	got, err := env.client.GetSecret(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, "modified", string(got.Data))
	assert.Equal(t, "v2", string(got.Metadata))
}

func TestIntegration_UpdateSecret_NotFound(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-update-nf", "pass")

	_, err := env.client.UpdateSecret("nonexistent-id", []byte("data"), []byte("meta"))
	assert.Error(t, err)
}

func TestIntegration_ListSecrets(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-list", "pass")

	secrets, err := env.client.ListSecrets()
	require.NoError(t, err)
	assert.Empty(t, secrets)

	env.client.CreateSecret(domain.SecretTypeLogin, []byte("cred1"), []byte("site1"))
	env.client.CreateSecret(domain.SecretTypeText, []byte("note1"), []byte("memo"))
	env.client.CreateSecret(domain.SecretTypeCard, []byte("card1"), nil)

	secrets, err = env.client.ListSecrets()
	require.NoError(t, err)
	assert.Len(t, secrets, 3)
}

func TestIntegration_GetSecret(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-get", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("hello"), []byte("greeting"))
	require.NoError(t, err)

	got, err := env.client.GetSecret(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, secret.ID, got.ID)
	assert.Equal(t, domain.SecretTypeText, got.Type)
	assert.Equal(t, "hello", string(got.Data))
	assert.Equal(t, "greeting", string(got.Metadata))
}

func TestIntegration_GetSecret_NotFound(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-get-nf", "pass")

	_, err := env.client.GetSecret("nonexistent-id")
	assert.Error(t, err)
}

func TestIntegration_DownloadBinary(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-dl", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeBinary, nil, []byte("archive.zip"))
	require.NoError(t, err)

	payload := []byte("PK\x03\x04fake-zip-content")
	err = env.client.UploadBinary(secret.ID, bytes.NewReader(payload), int64(len(payload)))
	require.NoError(t, err)

	data, err := env.client.DownloadBinary(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, payload, data)
}

func TestIntegration_DownloadBinary_NotFound(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-dl-nf", "pass")

	_, err := env.client.DownloadBinary("nonexistent-id")
	assert.Error(t, err)
}

func TestIntegration_DeleteSecret(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-delete", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("to-delete"), nil)
	require.NoError(t, err)

	err = env.client.DeleteSecret(secret.ID)
	require.NoError(t, err)

	_, err = env.client.GetSecret(secret.ID)
	assert.Error(t, err, "deleted secret should not be retrievable")
}

func TestIntegration_DeleteSecret_NotFound(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-delete-nf", "pass")

	err := env.client.DeleteSecret("nonexistent-id")
	assert.Error(t, err)
}

func TestIntegration_DeleteBinary(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-del-bin", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeBinary, nil, []byte("temp-file"))
	require.NoError(t, err)

	payload := []byte("binary-data")
	err = env.client.UploadBinary(secret.ID, bytes.NewReader(payload), int64(len(payload)))
	require.NoError(t, err)

	err = env.client.DeleteBinary(secret.ID)
	require.NoError(t, err)

	_, err = env.client.DownloadBinary(secret.ID)
	assert.Error(t, err, "deleted binary should not be downloadable")
}

func TestIntegration_Sync(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-sync", "pass")

	_, err := env.client.CreateSecret(domain.SecretTypeLogin, []byte("old-cred"), []byte("old-meta"))
	require.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)
	cursor := time.Now().UTC()
	time.Sleep(1100 * time.Millisecond)

	s2, err := env.client.CreateSecret(domain.SecretTypeText, []byte("note"), nil)
	require.NoError(t, err)
	s3, err := env.client.CreateSecret(domain.SecretTypeCard, []byte("card-data"), nil)
	require.NoError(t, err)

	synced, err := env.client.SyncSecrets(cursor)
	require.NoError(t, err)
	assert.Len(t, synced, 2, "should see exactly the two new secrets")

	ids := map[string]bool{synced[0].ID: true, synced[1].ID: true}
	assert.True(t, ids[s2.ID])
	assert.True(t, ids[s3.ID])
}

func TestIntegration_Sync_DeletedAppears(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-sync-del", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("will-delete"), nil)
	require.NoError(t, err)

	cursor := time.Now().UTC()
	time.Sleep(10 * time.Millisecond)

	err = env.client.DeleteSecret(secret.ID)
	require.NoError(t, err)

	synced, err := env.client.SyncSecrets(cursor)
	require.NoError(t, err)
	require.Len(t, synced, 1)
	assert.Equal(t, secret.ID, synced[0].ID)
	assert.NotNil(t, synced[0].DeletedAt)
}

func TestIntegration_Sync_EmptyWhenNoChanges(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-sync-empty", "pass")

	_, err := env.client.CreateSecret(domain.SecretTypeText, []byte("old"), nil)
	require.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)
	cursor := time.Now().UTC()

	synced, err := env.client.SyncSecrets(cursor)
	require.NoError(t, err)
	assert.Empty(t, synced)
}

func TestIntegration_BinaryLifecycle(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-bin-lifecycle", "pass")

	secret, err := env.client.CreateSecret(domain.SecretTypeBinary, nil, []byte("document.pdf"))
	require.NoError(t, err)

	payload := []byte("%PDF-1.4 fake pdf content")
	err = env.client.UploadBinary(secret.ID, bytes.NewReader(payload), int64(len(payload)))
	require.NoError(t, err)

	downloaded, err := env.client.DownloadBinary(secret.ID)
	require.NoError(t, err)
	assert.Equal(t, payload, downloaded)

	err = env.client.DeleteBinary(secret.ID)
	require.NoError(t, err)

	_, err = env.client.DownloadBinary(secret.ID)
	assert.Error(t, err, "binary should not be downloadable after delete")

	_, err = env.client.GetSecret(secret.ID)
	assert.Error(t, err, "secret should be soft-deleted after binary delete")
}

func TestIntegration_Unauthorized(t *testing.T) {
	env := setupIntegration(t)

	_, err := env.client.ListSecrets()
	assert.Error(t, err, "unauthenticated request should fail")

	_, err = env.client.CreateSecret(domain.SecretTypeText, []byte("data"), nil)
	assert.Error(t, err, "unauthenticated create should fail")
}

func TestIntegration_InvalidSecretType(t *testing.T) {
	env := setupIntegration(t)
	registerAndAuth(t, env, "user-invalid-type", "pass")

	_, err := env.client.CreateSecret(domain.SecretType("unknown"), []byte("data"), nil)
	assert.Error(t, err)
}

func TestIntegration_UserIsolation(t *testing.T) {
	env := setupIntegration(t)

	token1, err := env.client.Register("alice", "pass1")
	require.NoError(t, err)
	env.client.SetToken(token1)

	secret, err := env.client.CreateSecret(domain.SecretTypeText, []byte("alice-secret"), nil)
	require.NoError(t, err)

	client2 := api.NewClient(env.baseURL)
	token2, err := client2.Register("bob", "pass2")
	require.NoError(t, err)
	client2.SetToken(token2)

	_, err = client2.GetSecret(secret.ID)
	assert.Error(t, err, "bob should not access alice's secret")

	err = client2.DeleteSecret(secret.ID)
	assert.Error(t, err, "bob should not delete alice's secret")

	bobSecrets, err := client2.ListSecrets()
	require.NoError(t, err)
	assert.Empty(t, bobSecrets, "bob should have no secrets")
}

// Legacy helper kept for backward compatibility with raw HTTP tests.
func registerIntegrationUser(t *testing.T, baseURL string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"login":    "integration-user",
		"password": "integration-password",
	})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/v1/register", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload.Token)
	return payload.Token
}
