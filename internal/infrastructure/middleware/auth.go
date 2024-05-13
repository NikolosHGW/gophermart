package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"github.com/golang-jwt/jwt/v4"
)

const fullTokenLength = 2

type userIDKey string

var contextKey userIDKey = "userID"

type AuthMiddleware struct {
	secretKey string
}

func (am *AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Необходима аутентификация", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != fullTokenLength {
			http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
		claims := &usecase.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(am.secretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Неверный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey, claims.UserID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: secretKey,
	}
}
