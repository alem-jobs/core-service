package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var secret = "secret_key"

type contextKey string

const (
	UserIDKey           contextKey = "user_id"
	OrganizationIDKey   contextKey = "organization_id"
	OrganizationTypeKey contextKey = "organization_type"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenString, err := extractToken(authHeader)
		if err != nil {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid user_id in token", http.StatusUnauthorized)
			return
		}
		userID := int64(userIDFloat)

		orgIDFloat, ok := claims["organization_id"].(float64)
		if !ok {
			http.Error(w, "Invalid organization_id in token", http.StatusUnauthorized)
			return
		}
		orgID := int64(orgIDFloat)

		orgType, ok := claims["organization_type"].(string)
		if !ok {
			http.Error(w, "Invalid organization_id in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, OrganizationIDKey, orgID)
		ctx = context.WithValue(ctx, OrganizationTypeKey, orgType)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", http.ErrNoLocation
	}
	return parts[1], nil
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrBodyNotAllowed
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, http.ErrBodyNotAllowed
	}

	return claims, nil
}

func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}

func GetOrganizationID(r *http.Request) (int64, bool) {
	orgID, ok := r.Context().Value(OrganizationIDKey).(int64)
	return orgID, ok
}

func GetOrganizationType(r *http.Request) (string, bool) {
	orgType, ok := r.Context().Value(OrganizationTypeKey).(string)
	return orgType, ok
}
