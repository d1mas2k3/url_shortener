package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_config "github.com/d1mas2k3/url_shortener/internal/core/config"
	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"
	core_http_middleware "github.com/d1mas2k3/url_shortener/internal/core/transport/http/middleware"
	core_http_server "github.com/d1mas2k3/url_shortener/internal/core/transport/http/server"
	links_memory_repository "github.com/d1mas2k3/url_shortener/internal/features/links/repository/memory"
	links_service "github.com/d1mas2k3/url_shortener/internal/features/links/service"
	links_transport_http "github.com/d1mas2k3/url_shortener/internal/features/links/transport/http"
	"go.uber.org/zap"
)

func main() {
	// Ловим SIGINT/SIGTERM — при Ctrl+C сервер мягко завершится
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	// Логгер
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Общий конфиг приложения (STORAGE, BASE_URL)
	appConfig := core_config.NewConfigMust()

	// Выбор хранилища по конфигу
	logger.Debug("initializing storage", zap.String("storage", appConfig.Storage))
	var repo links_service.LinksRepository
	switch appConfig.Storage {
	case "memory":
		repo = links_memory_repository.NewLinksMemoryRepository()
	case "postgres":
		logger.Fatal("postgres storage not implemented yet")
	default:
		logger.Fatal("unknown storage type", zap.String("storage", appConfig.Storage))
	}

	// Сервис
	linksService := links_service.NewLinksService(repo)

	// HTTP handler
	linksHandler := links_transport_http.NewLinksHTTPHandler(linksService, appConfig.BaseURL)

	// HTTP сервер с middleware
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	// Роутер
	apiVersionRouter := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(linksHandler.Routes()...)
	httpServer.RegisterAPIRoutes(apiVersionRouter)

	logger.Debug("starting HTTP server")
	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
