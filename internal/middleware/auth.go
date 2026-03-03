package middleware

import (
	"context"
	"log"
	"net/http"

	"link-shortener/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

func WithAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie("user_id")
			if err == nil {
				parsedUserID, valid := auth.ParseCookieValue(cookie.Value)
				if valid {
					userID = parsedUserID
				}
			}

			if userID == "" {
				newUserID, err := auth.GenerateUserID()
				if err != nil {
					log.Printf("Failed to generate user ID: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
				userID = newUserID

				cookieValue := auth.BuildCookieValue(userID)
				http.SetCookie(w, &http.Cookie{
					Name:  "user_id",
					Value: cookieValue,
					Path:  "/",
				})
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
