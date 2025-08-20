package middleware

import (
	"net/http"
	"strings"

	"ingestion-service/internal/auth"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth wraps a handler and requires valid JWT authentication
func (a *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from header
		token, err := auth.ExtractTokenFromHeader(r)
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := auth.WithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireRole wraps a handler and requires a specific role
func (a *AuthMiddleware) RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !claims.HasRole(role) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireAnyRole wraps a handler and requires any of the specified roles
func (a *AuthMiddleware) RequireAnyRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !claims.HasAnyRole(roles...) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireAccountAccess ensures the user can only access their own account
func (a *AuthMiddleware) RequireAccountAccess(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Extract account ID from request (could be from path, query, or body)
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			// Try to get from path
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) > 2 {
				accountID = pathParts[2]
			}
		}

		// If no account ID in request, use the one from JWT
		if accountID == "" {
			accountID = claims.AccountID
		}

		// Check if user has access to this account
		if claims.AccountID != accountID && !claims.HasRole("admin") {
			http.Error(w, "Access denied to account", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
