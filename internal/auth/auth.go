package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const SecretKey = "super-secret-key-for-signing"

func GenerateUserID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func Sign(userID string) string {
	h := hmac.New(sha256.New, []byte(SecretKey))
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

func Verify(userID, signature string) bool {
	expectedSignature := Sign(userID)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func BuildCookieValue(userID string) string {
	signature := Sign(userID)
	return fmt.Sprintf("%s:%s", userID, signature)
}

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
