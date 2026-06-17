package core_http_middleware

// Middleware - код, который выполняется между получением запроса и хендлером(эндпоинтом)

import "net/http"

// Создает новый тип "Middleware", который принимает http.Handler и возвращает http.Handler
// http.Handler - интерфейс, который умеет обрабатывать HTTP запросы
type Middleware func(http.Handler) http.Handler

// Берёт хендлер и список мидлварей, и оборачивает хендлер в них по порядку.
func ChainMiddleware(
	h http.Handler,
	m ...Middleware,
) http.Handler {
	if len(m) == 0 {
		return h
	}

	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h) // Вызови итую функцию из слайза, передай туда h, и результат запиши в h
	}

	return h
}
