package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/golang-jwt/jwt/v4"
)

const fullTokenLength = 2

type AuthMiddleware struct {
	secretKey string
}

func (am *AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != fullTokenLength {
			http.Error(w, "неверный формат токена", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
		userID := GetUserID(tokenString, am.secretKey)
		if userID == -1 {
			http.Error(w, "неверный токен", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), domain.ContextKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAuthMiddleware(secretKey string) *AuthMiddleware {
	return &AuthMiddleware{
		secretKey: secretKey,
	}
}

func GetUserID(tokenString string, secretKey string) int {
	claims := &entity.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		return -1
	}

	return claims.UserID
}
