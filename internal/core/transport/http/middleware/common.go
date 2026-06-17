package core_http_middleware

import (
	"net/http"
	"time"

	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"

	core_http_response "github.com/d1mas2k3/url_shortener/internal/core/transport/http/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	requestIDHeader = "X-Request-ID"
)

// Структура состоит из 3 этапов.
// В начале мы объявляем функцию, которая возвращает Middleware
// После этого, мы принимаем эстафету от предыдущего Middleware
// А дальше реализуем логику, упаковываем ее в тип http.HandlerFunc, и возвращаем
// func Func_name(...) Middleware {
//     return func(next http.Handler) http.Handler { -
//         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         ...
//         next.ServeHTTP(w, r) - передаем эстафету следующему middleware
//         }
//     }
// }

// Проставляем ID запроса если его не было
func RequestID() Middleware { // Middleware из middleware.go, так как одинаковый пакет
	return func(next http.Handler) http.Handler { // Возвращаемая функция должна быть типа http.Handler

		// Превращаем обычную функцию в http.Handler с помощью http.HandlerFunc
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Создаем ID запроса
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}

			r.Header.Set(requestIDHeader, requestID)   // Записываем id в запрос
			w.Header().Set(requestIDHeader, requestID) // Записываем id в ответ

			// Отправляем запрос на обработку следующему хендлеру после Middleware
			next.ServeHTTP(w, r)
		})
	}
}

// Если предыдущий Middleware добавлял ID запроса, то этот добавляет URL и помещает все в контекст
func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Получаем ID запроса из предыдущей функции
			requestID := r.Header.Get(requestIDHeader)

			// Создаёт новый логгер с прикреплёнными полями
			l := log.With(
				zap.String("request_id", requestID), // Поле с ID
				zap.String("url", r.URL.String()),   // Поле с URL (например /products)
			)

			// Кладёт этот логгер в контекст запроса под ключом "log"
			ctx := core_logger.ToContext(r.Context(), l)

			// Передаёт запрос дальше, но уже с новым контекстом где лежит логгер
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Логирует каждый запрос — когда пришёл и сколько времени занял
func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := core_logger.FromContext(r.Context())

			// Оборачиваем стандартный w в наш ResponseWriter который умеет запоминать статус код
			rw := core_http_response.NewResponseWriter(w)

			// Запоминаем время ДО выполнения запроса и логируем что запрос пришёл
			before := time.Now()
			log.Debug(
				">>> incoming HTTP request",
				zap.String("http_method", r.Method),
				zap.Time("time", before.UTC()),
			)
			next.ServeHTTP(rw, r)

			// После того как хендлер отработал — логируем результат.
			log.Debug(
				"<<< done HTTP request",
				zap.Int("status_code", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
			)
		})
	}
}

// Middleware которая ловит паники и возвращает клиенту 500 вместо падения сервера
func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := core_logger.FromContext(r.Context())                          // Достаем контекст запроса
			responseHandler := core_http_response.NewHTTPResponseHandler(log, w) // w для отправки ответа

			// Ошибка на стороне сервера; подробнее про PanicResponse в handler.go
			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(
						p,
						"during handle HTTP request go unexpected panic",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
