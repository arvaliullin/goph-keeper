package domain

import "errors"

var (
	// ErrInvalidInput возвращается при некорректных входных данных.
	ErrInvalidInput = errors.New("invalid input")
	// ErrInvalidCredentials возвращается при ошибке аутентификации.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserAlreadyExists возвращается при попытке зарегистрировать существующего пользователя.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserNotFound возвращается, когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidSecretType возвращается при передаче неподдерживаемого типа секрета.
	ErrInvalidSecretType = errors.New("invalid secret type")
	// ErrSecretNotFound возвращается, когда секрет не найден.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrBinarySecretRequired возвращается при попытке работать с бинарным API для небинарного секрета.
	ErrBinarySecretRequired = errors.New("binary secret required")
)
