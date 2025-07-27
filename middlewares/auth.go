package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "User ID not found in token", http.StatusUnauthorized)
			return
		}
		userID := int(userIDFloat)
		role := claims["role"].(string)

		// Pass userID via context
		ctx := context.WithValue(r.Context(), userKey, userID)
		ctx = context.WithValue(ctx, roleKey, role)

		next(w, r.WithContext(ctx))
	}
}

// Helper to retrieve user ID in handler
func GetUserID(r *http.Request) int {
	if userID, ok := r.Context().Value(userKey).(int); ok {
		return userID
	}
	return 0
}

func GetUserRole(r *http.Request) string {
	if role, ok := r.Context().Value(roleKey).(string); ok {
		return role
	}
	return ""
}

// prevent non-admins from accessing certain routes
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := GetUserRole(r)
		if role != "admin" {
			http.Error(w, "Admins only", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
