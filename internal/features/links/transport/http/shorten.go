package links_transport_http

import (
	"net/http"

	core_logger "github.com/d1mas2k3/url_shortener/internal/core/logger"
	core_http_request "github.com/d1mas2k3/url_shortener/internal/core/transport/http/request"
	core_http_response "github.com/d1mas2k3/url_shortener/internal/core/transport/http/response"
)

// HTTP-обёртка над бизнес-логикой создания короткой ссылки. Принимает POST-запрос с
// оригинальным URL, отдаёт его в сервис, возвращает клиенту короткий URL в JSON.
func (h *LinksHTTPHandler) Shorten(rw http.ResponseWriter, r *http.Request) {
	// Достаём контекст запроса и логгер, который middleware положил в контекст
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	log.Debug("invoke shorten handler")

	// Создаём responseHandler - он умеет отправлять JSON и маппить ошибки в HTTP-статусы
	responseHandler := core_http_response.NewHTTPResponseHandler(log, rw)

	// Декодируем JSON из тела запроса в структуру ShortenRequest и валидируем по тегам
	var request ShortenRequest
	if err := core_http_request.DecodeAnyValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode request")
		return
	}

	// Передаём оригинальный URL в сервис - он вернёт domain.Link с кодом
	link, err := h.linksService.Shorten(ctx, request.URL)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to shorten url")
		return
	}

	// Маппим domain.Link в DTO и отправляем клиенту 201 Created
	response := linkDTOFromDomain(link, h.baseURL)
	responseHandler.JSONResponse(response, http.StatusCreated)
}
