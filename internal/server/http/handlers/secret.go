package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/arvaliullin/goph-keeper/internal/core/domain"
	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
)

// SecretHandler обрабатывает запросы управления секретами.
type SecretHandler struct {
	service ports.SecretService
}

// NewSecretHandler создает новый SecretHandler.
func NewSecretHandler(service ports.SecretService) *SecretHandler {
	return &SecretHandler{service: service}
}

type secretRequest struct {
	Type     domain.SecretType `json:"type"`
	Data     []byte            `json:"data"`
	Metadata []byte            `json:"metadata"`
}

// Create создает новый секрет.
// @Summary Создать секрет
// @Description Создает новый секрет (пароль, текст, карту и т.д.)
// @Tags secrets
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body secretRequest true "Данные секрета"
// @Success 200 {object} domain.Secret
// @Router /api/v1/secrets [post]
func (h *SecretHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	secret, err := h.service.Create(r.Context(), userID, req.Type, req.Data, req.Metadata)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Update обновляет существующий секрет.
// @Summary Обновить секрет
// @Description Обновляет данные и метаданные существующего секрета
// @Tags secrets
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "ID секрета"
// @Param request body secretRequest true "Новые данные секрета"
// @Success 200 {object} domain.Secret
// @Router /api/v1/secrets/{id} [put]
func (h *SecretHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req secretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	secret, err := h.service.Update(r.Context(), userID, id, req.Data, req.Metadata)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// List возвращает список секретов или инкрементальные изменения.
// @Summary Получить список секретов
// @Description Возвращает все секреты текущего пользователя или изменения после updated_after
// @Tags secrets
// @Security ApiKeyAuth
// @Produce json
// @Param updated_after query string false "RFC3339 timestamp for incremental sync"
// @Success 200 {array} domain.Secret
// @Router /api/v1/secrets [get]
func (h *SecretHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		secrets []*domain.Secret
		err     error
	)

	updatedAfterRaw := r.URL.Query().Get("updated_after")
	if updatedAfterRaw != "" {
		updatedAfter, parseErr := time.Parse(time.RFC3339, updatedAfterRaw)
		if parseErr != nil {
			http.Error(w, "invalid updated_after format", http.StatusBadRequest)
			return
		}
		secrets, err = h.service.Sync(r.Context(), userID, updatedAfter)
	} else {
		secrets, err = h.service.List(r.Context(), userID)
	}
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secrets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Get возвращает секрет по идентификатору.
// @Summary Получить секрет
// @Description Возвращает данные секрета по ID
// @Tags secrets
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "ID секрета"
// @Success 200 {object} domain.Secret
// @Router /api/v1/secrets/{id} [get]
func (h *SecretHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	secret, err := h.service.Get(r.Context(), userID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Delete удаляет секрет.
// @Summary Удалить секрет
// @Description Удаляет секрет по ID
// @Tags secrets
// @Security ApiKeyAuth
// @Param id path string true "ID секрета"
// @Success 204
// @Router /api/v1/secrets/{id} [delete]
func (h *SecretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	err := h.service.Delete(r.Context(), userID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
