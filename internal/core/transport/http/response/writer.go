package core_http_response

import "net/http"

// "Расширяем" обычный http.ResponseWriter, чтобы он смог возвращать statusCode
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

var (
	StatusCodeUnitialized = -1 // Маркер, что статус код не установлен
)

// Конструктор, оборачивает стандартный ResponseWriter в наш, статус изначально -1
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     StatusCodeUnitialized,
	}
}

// Переопределяем WriteHeader. Когда хендлер отправляет статус-код, мы сохраняем его в наш statusCode
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

// Возвращает сохранённый статус код. Если WriteHeader ещё не вызывался — паникует.
func (rw *ResponseWriter) GetStatusCode() int {
	if rw.statusCode == StatusCodeUnitialized {
		return http.StatusOK
	}
	return rw.statusCode
}
