// Гнерирует случайный код из 10 символов. Утилита без зависимостей
package links_service

import (
	"crypto/rand"
	"fmt"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	codeLength = 10
)

func generateCode() (string, error) {
	b := make([]byte, codeLength)

	// Заполняем случайными байтами наш слайс от 0 до 255 [0, 43, 198 и тп]
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	// 255 % 63(длина алфавита). Другими словами, перезаписываем числа в нашем диапазоне (0 до 63 (по тз))
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}

	return string(b), nil
}
