package core_http_response
// Пакет отвечает за формирование и отправку HTTP-ответов клиенту
// Содержит методы для успешных ответов (JSON, NoContent),
// обработки ошибок с маппингом на HTTP статус коды, и обработки паник с возвратом 500

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"
	"go.uber.org/zap"
)

// Структура которая хранит два инструмента: логгер и объект для записи ответа клиенту
type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(
	log *core_logger.Logger,
	rw http.ResponseWriter,
) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

// Отправляет JSON-ответ клиенту
func (h *HTTPResponseHandler) JSONResponse(
	responseBody any,
	stausCode int,
) {
	h.rw.WriteHeader(stausCode) // Устанавливает HTTP статус код (200, 400, 500 и тд)

	// Берёт любую Go-структуру, превращает её в JSON и пишет в ответ
	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("write HTTP response", zap.Error(err))
	}
}

// Когда операция выполнилась успешно и возвращать пользователю ничего не нужно
func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFunc    func(string, ...zap.Field)
	)

	// Обрабатывает statusCode ошибок
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn

	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug

	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))
	h.errorResponse(statusCode, err, msg)
}

// Это метод логирует ошибку когда в приложении случилась паника.
func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError

	// Создаёт ошибку из значения паники
	err := fmt.Errorf("unexpected panic : %v", p) // %v - конвертирует все что угодно в строку

	h.log.Error(msg, zap.Error(err)) // 1. Логирует ошибку

	h.errorResponse(statusCode, err, msg)
}

// Отправляет клиенту статус 500, формирует json ответ и отправляет клиенту
func (h *HTTPResponseHandler) errorResponse(
	statusCode int,
	err error,
	msg string,
) {
	h.rw.WriteHeader(statusCode) // 2. Отправляет клиенту статус 500

	response := map[string]string{ // 3. Формирует JSON ответ и отправляет клиенту
		"message": msg,
		"error":   err.Error(),
	}

	// Отправляет клиенту ответ в формате json; если не получилось отправить - возвращает ошибку
	h.JSONResponse(
		response,
		statusCode,
	)
}
