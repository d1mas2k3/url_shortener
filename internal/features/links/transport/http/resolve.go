package links_transport_http

import (
	"net/http"

	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"
	core_http_response "github.com/d1mas2k3/url_shortener/internal/core/transport/http/response"
)

// HTTP-обёртка над поиском по коду. Принимает GET-запрос с кодом в пути (/links/abc1234567), 
// отдаёт в сервис, возвращает клиенту оригинальный URL в JSON.
func (h *LinksHTTPHandler) Resolve(rw http.ResponseWriter, r *http.Request) {
	// Достаём контекст запроса и логгер, который middleware положил в контекст
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	log.Debug("invoke resolve handler")

	// Создаём responseHandler - он умеет отправлять JSON и маппить ошибки в HTTP-статусы
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	// Извлекаем значение {code} из URL-пути, например /r/abc1234567 → "abc1234567"
	code := r.PathValue("code")

	// Передаём code в сервис - он вернёт domain.Link с оригинальным URL
	link, err := h.linksService.Resolve(ctx, code)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to resolve code")
		return
	}

	// Маппим domain.Link в DTO и отправляем клиенту 200 OK
	response := ResolveResponse{OriginalURL: link.OriginalURL}
	responseHandler.JSONResponse(response, http.StatusOK)
}
