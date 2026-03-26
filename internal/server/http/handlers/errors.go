package handlers

import (
	"errors"
	"net/http"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
)

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrInvalidSecretType), errors.Is(err, domain.ErrBinarySecretRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrUserAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrSecretNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
