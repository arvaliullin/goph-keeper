package domain

import "time"

// SecretType тип хранимого секрета.
type SecretType string

const (
	// SecretTypeLogin тип для пары логин/пароль.
	SecretTypeLogin SecretType = "login"
	// SecretTypeText тип для произвольных текстовых данных.
	SecretTypeText SecretType = "text"
	// SecretTypeBinary тип для бинарных данных.
	SecretTypeBinary SecretType = "binary"
	// SecretTypeCard тип для данных банковской карты.
	SecretTypeCard SecretType = "card"
)

// Valid проверяет, поддерживается ли тип секрета.
func (t SecretType) Valid() bool {
	switch t {
	case SecretTypeLogin, SecretTypeText, SecretTypeBinary, SecretTypeCard:
		return true
	default:
		return false
	}
}

// Secret представляет собой хранимые пользователем приватные данные.
type Secret struct {
	// ID уникальный идентификатор секрета (UUID).
	ID string `json:"id"`
	// UserID идентификатор владельца секрета.
	UserID int64 `json:"user_id"`
	// Type тип секрета (login, text, binary, card).
	Type SecretType `json:"type"`
	// Data зашифрованные данные секрета.
	Data []byte `json:"data"`
	// Metadata зашифрованные метаданные секрета.
	Metadata []byte `json:"metadata"`
	// CreatedAt время создания секрета.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt время последнего обновления секрета.
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt время мягкого удаления (nil если не удален).
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
