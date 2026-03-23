package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/arvaliullin/goph-keeper/internal/core/ports"
)

// AuthHandler обрабатывает запросы аутентификации и регистрации.
type AuthHandler struct {
	service ports.AuthService
}

// NewAuthHandler создает новый AuthHandler.
func NewAuthHandler(service ports.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type authRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Register регистрирует нового пользователя.
// @Summary Регистрация пользователя
// @Description Регистрация нового пользователя по логину и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Данные для регистрации"
// @Success 200 {object} authResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 409 {string} string "Conflict"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse{Token: token}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Login аутентифицирует пользователя.
// @Summary Аутентификация пользователя
// @Description Аутентификация пользователя и получение JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authRequest true "Учетные данные"
// @Success 200 {object} authResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/v1/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(authResponse{Token: token}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
