package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/go-resty/resty/v2"
)

// Client клиент для взаимодействия с GophKeeper API.
type Client struct {
	client *resty.Client
	token  string
}

// NewClient создает новый экземпляр API клиента.
func NewClient(baseURL string) *Client {
	return &Client{
		client: resty.New().SetBaseURL(baseURL),
	}
}

// SetToken устанавливает JWT токен для последующих запросов.
func (c *Client) SetToken(token string) {
	c.token = token
	c.client.SetAuthToken(token)
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Register регистрирует пользователя и возвращает токен.
func (c *Client) Register(login, password string) (string, error) {
	var resp authResponse
	res, err := c.client.R().
		SetBody(authRequest{Login: login, Password: password}).
		SetResult(&resp).
		Post("/api/v1/register")

	if err != nil {
		return "", err
	}
	if res.IsError() {
		return "", fmt.Errorf("register failed: %s", res.String())
	}

	return resp.Token, nil
}

// Login аутентифицирует пользователя и возвращает токен.
func (c *Client) Login(login, password string) (string, error) {
	var resp authResponse
	res, err := c.client.R().
		SetBody(authRequest{Login: login, Password: password}).
		SetResult(&resp).
		Post("/api/v1/login")

	if err != nil {
		return "", err
	}
	if res.IsError() {
		return "", fmt.Errorf("login failed: %s", res.String())
	}

	return resp.Token, nil
}

type secretRequest struct {
	Type     domain.SecretType `json:"type"`
	Data     []byte            `json:"data"`
	Metadata []byte            `json:"metadata"`
}

// CreateSecret создает новый секрет.
func (c *Client) CreateSecret(typ domain.SecretType, data, metadata []byte) (*domain.Secret, error) {
	var secret domain.Secret
	res, err := c.client.R().
		SetBody(secretRequest{Type: typ, Data: data, Metadata: metadata}).
		SetResult(&secret).
		Post("/api/v1/secrets")

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("create secret failed: %s", res.String())
	}

	return &secret, nil
}

// ListSecrets получает список секретов пользователя.
func (c *Client) ListSecrets() ([]*domain.Secret, error) {
	var secrets []*domain.Secret
	res, err := c.client.R().
		SetResult(&secrets).
		Get("/api/v1/secrets")

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("list secrets failed: %s", res.String())
	}

	return secrets, nil
}

// SyncSecrets получает изменения секретов после указанного времени.
func (c *Client) SyncSecrets(updatedAfter time.Time) ([]*domain.Secret, error) {
	var secrets []*domain.Secret
	res, err := c.client.R().
		SetResult(&secrets).
		SetQueryParam("updated_after", updatedAfter.UTC().Format(time.RFC3339)).
		Get("/api/v1/secrets")

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("sync secrets failed: %s", res.String())
	}

	return secrets, nil
}

// GetSecret получает секрет по ID.
func (c *Client) GetSecret(id string) (*domain.Secret, error) {
	var secret domain.Secret
	res, err := c.client.R().
		SetResult(&secret).
		Get("/api/v1/secrets/" + id)

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("get secret failed: %s", res.String())
	}

	return &secret, nil
}

// UpdateSecret обновляет данные и метаданные секрета.
func (c *Client) UpdateSecret(id string, data, metadata []byte) (*domain.Secret, error) {
	var secret domain.Secret
	res, err := c.client.R().
		SetBody(secretRequest{Data: data, Metadata: metadata}).
		SetResult(&secret).
		Put("/api/v1/secrets/" + id)

	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("update secret failed: %s", res.String())
	}

	return &secret, nil
}

// DeleteSecret удаляет секрет.
func (c *Client) DeleteSecret(id string) error {
	res, err := c.client.R().Delete("/api/v1/secrets/" + id)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("delete secret failed: %s", res.String())
	}
	return nil
}

// UploadBinary загружает бинарный файл.
func (c *Client) UploadBinary(id string, reader io.Reader, size int64) error {
	req, err := http.NewRequest("POST", c.client.BaseURL+"/api/v1/binary/"+url.PathEscape(id), reader)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload binary failed: %s", string(body))
	}

	return nil
}

// DownloadBinary скачивает бинарный файл.
func (c *Client) DownloadBinary(id string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.client.BaseURL+"/api/v1/binary/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download binary failed: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}

// DeleteBinary удаляет бинарный секрет.
func (c *Client) DeleteBinary(id string) error {
	req, err := http.NewRequest("DELETE", c.client.BaseURL+"/api/v1/binary/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		return fmt.Errorf("delete binary failed: %s", string(body))
	}

	return nil
}
