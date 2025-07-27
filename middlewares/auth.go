package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	userKey contextKey = "userID"
	roleKey contextKey = "role"
)

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "User ID not found in token", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
			return
		}

		role := claims["role"].(string)

		// Store UUID and role in context
		ctx := context.WithValue(r.Context(), userKey, userID)
		ctx = context.WithValue(ctx, roleKey, role)

		next(w, r.WithContext(ctx))
	}
}

// ✅ UUID-based
func GetUserID(r *http.Request) uuid.UUID {
	if userID, ok := r.Context().Value(userKey).(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

func GetUserRole(r *http.Request) string {
	if role, ok := r.Context().Value(roleKey).(string); ok {
		return role
	}
	return ""
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GetUserRole(r) != "admin" {
			http.Error(w, "Admins only", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
