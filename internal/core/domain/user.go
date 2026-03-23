package domain

import "time"

// User представляет пользователя системы.
type User struct {
	// ID уникальный идентификатор пользователя.
	ID int64 `json:"id"`
	// Login логин пользователя.
	Login string `json:"login"`
	// PasswordHash bcrypt-хеш пароля.
	PasswordHash string `json:"-"`
	// CreatedAt время регистрации пользователя.
	CreatedAt time.Time `json:"created_at"`
}
