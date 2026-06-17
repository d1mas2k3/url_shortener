package core_http_server

// Его задача одна: запустить HTTP-сервер и корректно его остановить

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"
	core_http_middleware "github.com/d1mas2k3/url_shortener/internal/core/transport/http/middleware"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux        *http.ServeMux // Какой обработчик на какой URL вызвать
	config     Config         // Настройка сервера
	log        *core_logger.Logger
	middleware []core_http_middleware.Middleware
}

// Его задача одна: запустить HTTP-сервер и корректно его остановить
func NewHTTPServer(
	config Config,
	log *core_logger.Logger,
	middleware ...core_http_middleware.Middleware,
) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		config:     config,
		log:        log,
		middleware: middleware,
	}
}

func (s *HTTPServer) RegisterAPIRoutes(routers ...*APIVersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.apiVersion)

		s.mux.Handle(
			prefix+"/",
			http.StripPrefix(prefix, router.WithMiddleware()),
		)
	}
}

// Сердце сервера, делает три вещи по очереди:
func (s *HTTPServer) Run(ctx context.Context) error {
	mux := core_http_middleware.ChainMiddleware(s.mux, s.middleware...)

	server := &http.Server{
		Addr:    s.config.Addr, // На каком порту слушать запросы
		Handler: mux,           // Кто обрабатывает HTTP-запросы (endpoint -> handler)
	}

	ch := make(chan error, 1)

	// 1. Запускает сервер в горутине
	go func() {
		defer close(ch)
		// Запись лога уровня Warn
		s.log.Warn("start HTTP server", zap.String("addr", s.config.Addr))

		err := server.ListenAndServe() // Висит и вечно слушает запросы

		// Проверяет какой кейс из select ниже выполнился. Если ошибка, то возвращает ее
		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	// 2. Ждёт одно из двух событий
	select {
	case err := <-ch: // Сервер упал сам
		if err != nil {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
	case <-ctx.Done(): // Пришёл сигнал завершения (Ctrl+C)
		s.log.Warn("shutdown HTTP server")

		// 3. Graceful shutdown — мягкое выключение (когда приходит <-ctx.Done())
		shutdownCtx, cancel := context.WithTimeout( // Контекст с таймаутом
			context.Background(),
			s.config.ShutdownTimeout, // Переменная из пакета config.go
		)
		defer cancel()

		// Если запрос не успел выполниться за отведенное время, останавливаем сервер
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		s.log.Warn("HTTP server stopped")
	}
	return nil
}
