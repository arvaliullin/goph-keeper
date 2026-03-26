package api

import (
	"fmt"
	"io"
	"iter"
	"net/url"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/go-resty/resty/v2"
)

// Client клиент для взаимодействия с GophKeeper API.
type Client struct {
	client *resty.Client
}

// NewClient создает новый экземпляр API клиента.
func NewClient(baseURL string) *Client {
	return &Client{
		client: resty.New().SetBaseURL(baseURL),
	}
}

// SetToken устанавливает JWT токен для последующих запросов.
func (c *Client) SetToken(token string) {
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

// SecretsIter возвращает итератор по секретам пользователя.
func (c *Client) SecretsIter() iter.Seq2[*domain.Secret, error] {
	return func(yield func(*domain.Secret, error) bool) {
		secrets, err := c.ListSecrets()
		if err != nil {
			yield(nil, err)
			return
		}
		for _, s := range secrets {
			if !yield(s, nil) {
				return
			}
		}
	}
}

// UploadBinary загружает бинарный файл.
func (c *Client) UploadBinary(id string, reader io.Reader, size int64) error {
	resp, err := c.client.R().
		SetBody(reader).
		SetContentLength(true).
		SetHeader("Content-Type", "application/octet-stream").
		Post("/api/v1/binary/" + url.PathEscape(id))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("upload binary failed: %s", resp.String())
	}
	return nil
}

// DownloadBinary скачивает бинарный файл.
func (c *Client) DownloadBinary(id string) ([]byte, error) {
	resp, err := c.client.R().
		SetDoNotParseResponse(true).
		Get("/api/v1/binary/" + url.PathEscape(id))
	if err != nil {
		return nil, err
	}
	defer resp.RawBody().Close()
	if resp.IsError() {
		body, _ := io.ReadAll(resp.RawBody())
		return nil, fmt.Errorf("download binary failed: %s", string(body))
	}
	return io.ReadAll(resp.RawBody())
}

// DeleteBinary удаляет бинарный секрет.
func (c *Client) DeleteBinary(id string) error {
	resp, err := c.client.R().
		Delete("/api/v1/binary/" + url.PathEscape(id))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("delete binary failed: %s", resp.String())
	}
	return nil
}
