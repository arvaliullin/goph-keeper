package handlers

import (
	"io"
	"net/http"

	"github.com/arvaliullin/goph-keeper/internal/core/ports"
	"github.com/arvaliullin/goph-keeper/internal/server/http/middleware"
	"github.com/go-chi/chi/v5"
)

// BinaryHandler обрабатывает запросы загрузки и скачивания бинарных данных.
type BinaryHandler struct {
	service       ports.BinaryService
	maxBinarySize int64
}

// NewBinaryHandler создает новый BinaryHandler.
func NewBinaryHandler(service ports.BinaryService, maxBinarySize int64) *BinaryHandler {
	return &BinaryHandler{service: service, maxBinarySize: maxBinarySize}
}

// Upload загружает бинарные данные.
// @Summary Загрузить бинарный файл
// @Description Загружает зашифрованный бинарный payload для бинарного секрета
// @Tags binary
// @Security ApiKeyAuth
// @Accept application/octet-stream
// @Param id path string true "ID файла"
// @Success 200
// @Router /api/v1/binary/{id} [post]
func (h *BinaryHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	size := r.ContentLength
	if size <= 0 {
		http.Error(w, "missing content length", http.StatusBadRequest)
		return
	}
	if size > h.maxBinarySize {
		http.Error(w, "binary payload is too large", http.StatusRequestEntityTooLarge)
		return
	}

	body := http.MaxBytesReader(w, r.Body, h.maxBinarySize)
	defer body.Close()

	err := h.service.Upload(r.Context(), userID, id, body, size)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Download скачивает бинарные данные.
// @Summary Скачать бинарный файл
// @Description Скачивает и расшифровывает бинарный payload
// @Tags binary
// @Security ApiKeyAuth
// @Produce application/octet-stream
// @Param id path string true "ID файла"
// @Success 200 {string} string "Binary data"
// @Router /api/v1/binary/{id} [get]
func (h *BinaryHandler) Download(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	reader, err := h.service.Download(r.Context(), userID, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "failed to send file", http.StatusInternalServerError)
	}
}

// Delete удаляет бинарный секрет.
// @Summary Удалить бинарный файл
// @Description Удаляет бинарный файл и tombstone запись секрета
// @Tags binary
// @Security ApiKeyAuth
// @Param id path string true "ID файла"
// @Success 204
// @Router /api/v1/binary/{id} [delete]
func (h *BinaryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
