package core_http_request

// Парсит JSON и валидирует его

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "github.com/d1mas2k3/url_shortener/internal/core/errors"
	"github.com/go-playground/validator/v10"
)

var requestValidator = validator.New() // Создаётся один валидатор на весь пакет

func DecodeAnyValidateRequest(r *http.Request, dest any) error {
    // Читает тело запроса и пишет данные в dest
    if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
        return fmt.Errorf(
            "decode json: %v: %w", 
            err, 
            core_errors.ErrInvalidArgument,
        )
    }

    // Проверяет структуру по тегам validate на полях:
    if err := requestValidator.Struct(dest); err != nil {
        return fmt.Errorf(
            "request validator: %v: %w", 
            err,
            core_errors.ErrInvalidArgument,
        ) 
    }

    return nil
}