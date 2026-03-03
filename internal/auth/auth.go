// Package auth отвечает за аутентификацию пользователей через cookie.
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// SecretKey — ключ для подписи cookie.
const SecretKey = "super-secret-key-for-signing"

// GenerateUserID генерирует случайный идентификатор пользователя.
func GenerateUserID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Sign создает подпись для переданного userID.
func Sign(userID string) string {
	h := hmac.New(sha256.New, []byte(SecretKey))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify проверяет, что подпись соответствует userID.
func Verify(userID, signature string) bool {
	expectedSignature := Sign(userID)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// BuildCookieValue собирает значение cookie в формате "userID:signature".
func BuildCookieValue(userID string) string {
	signature := Sign(userID)
	return fmt.Sprintf("%s:%s", userID, signature)
}

// ParseCookieValue разбирает значение cookie и проверяет подпись.
// Возвращает userID и флаг валидности.
func ParseCookieValue(cookieValue string) (userID string, valid bool) {
	parts := strings.Split(cookieValue, ":")
	if len(parts) != 2 {
		return "", false
	}

	userID = parts[0]
	signature := parts[1]

	if !Verify(userID, signature) {
		return "", false
	}

	return userID, true
}
