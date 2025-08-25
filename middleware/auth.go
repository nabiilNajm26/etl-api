package middleware

import (
	"context"
	"etl-api/utils"
	"net/http"
	"strings"
)

// JWTAuth validates JWT tokens and adds user info to request context
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error": "Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, `{"error": "Token is required"}`, http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmailFromContext extracts user email from request context
func GetUserEmailFromContext(r *http.Request) string {
	email, ok := r.Context().Value("user_email").(string)
	if !ok {
		return ""
	}
	return email
}