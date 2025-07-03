package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/iudx-sandbox-backend/pkg/authutility"
	"github.com/iudx-sandbox-backend/pkg/logger"
	"github.com/julienschmidt/httprouter"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// AuthorizeRequest validates JWT token and ensures user is authenticated
func AuthorizeRequest(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		requestID := getRequestID(ctx)

		logger.DebugWithContext(ctx, "Starting JWT token validation")

		err := authutility.TokenValid(r)
		if err != nil {
			logger.SecurityLog(ctx, "jwt_authentication_failed", false, map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"path":        r.URL.Path,
				"method":      r.Method,
				"user_agent":  r.UserAgent(),
				"error":       err.Error(),
				"request_id":  requestID,
			})

			logger.WarnWithMetadata("Authentication failed - invalid JWT token", map[string]interface{}{
				"request_id": requestID,
				"path":       r.URL.Path,
				"error":      err.Error(),
			})

			sendErrorResponse(w, http.StatusUnauthorized, "authentication_required", "Valid authentication token is required")
			return
		}

		// Extract user information for context
		user, err := authutility.ExtractTokenMetadata(r)
		if err != nil {
			logger.ErrorWithMetadata("Token extraction failed after validation", map[string]interface{}{
				"request_id": requestID,
				"path":       r.URL.Path,
			}, err)

			sendErrorResponse(w, http.StatusUnauthorized, "token_extraction_failed", "Failed to extract user information from token")
			return
		}

		// Add user to context for downstream handlers
		ctx = context.WithValue(ctx, "user_id", user.UserName)
		r = r.WithContext(ctx)

		logger.SecurityLog(ctx, "jwt_authentication_success", true, map[string]interface{}{
			"username":   user.UserName,
			"user_id":    user.UserID,
			"roles":      user.Roles,
			"path":       r.URL.Path,
			"method":     r.Method,
			"request_id": requestID,
		})

		logger.DebugWithContext(ctx, "JWT authentication successful")

		next(w, r, ps)
	}
}

// AuthorizeWithRoles validates JWT token and checks if user has required roles
func AuthorizeWithRoles(requiredRoles ...string) func(httprouter.Handle) httprouter.Handle {
	return func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			ctx := r.Context()
			requestID := getRequestID(ctx)

			logger.DebugWithMetadata("Starting role-based authorization", map[string]interface{}{
				"required_roles": requiredRoles,
				"path":           r.URL.Path,
				"request_id":     requestID,
			})

			// First validate token
			err := authutility.TokenValid(r)
			if err != nil {
				logger.SecurityLog(ctx, "role_auth_failed_invalid_token", false, map[string]interface{}{
					"remote_addr":    r.RemoteAddr,
					"path":           r.URL.Path,
					"method":         r.Method,
					"required_roles": requiredRoles,
					"error":          err.Error(),
					"request_id":     requestID,
				})

				sendErrorResponse(w, http.StatusUnauthorized, "authentication_required", "Valid authentication token is required")
				return
			}

			// Extract user information
			user, err := authutility.ExtractTokenMetadata(r)
			if err != nil {
				logger.ErrorWithMetadata("Token extraction failed during role authorization", map[string]interface{}{
					"request_id":     requestID,
					"required_roles": requiredRoles,
				}, err)

				sendErrorResponse(w, http.StatusUnauthorized, "token_extraction_failed", "Failed to extract user information from token")
				return
			}

			// Add user to context
			ctx = context.WithValue(ctx, "user_id", user.UserName)
			r = r.WithContext(ctx)

			// Check if user has any of the required roles
			if !user.HasAnyRole(requiredRoles...) {
				logger.SecurityLog(ctx, "role_authorization_failed", false, map[string]interface{}{
					"username":       user.UserName,
					"user_id":        user.UserID,
					"user_roles":     user.Roles,
					"required_roles": requiredRoles,
					"path":           r.URL.Path,
					"method":         r.Method,
					"request_id":     requestID,
				})

				logger.WarnWithMetadata("Authorization failed - insufficient roles", map[string]interface{}{
					"username":       user.UserName,
					"user_roles":     user.Roles,
					"required_roles": requiredRoles,
					"path":           r.URL.Path,
					"request_id":     requestID,
				})

				sendErrorResponse(w, http.StatusForbidden, "insufficient_privileges",
					fmt.Sprintf("Access denied. Required roles: %v", requiredRoles))
				return
			}

			logger.SecurityLog(ctx, "role_authorization_success", true, map[string]interface{}{
				"username":       user.UserName,
				"user_id":        user.UserID,
				"user_roles":     user.Roles,
				"required_roles": requiredRoles,
				"path":           r.URL.Path,
				"method":         r.Method,
				"request_id":     requestID,
			})

			logger.DebugWithMetadata("Role-based authorization successful", map[string]interface{}{
				"username":       user.UserName,
				"granted_roles":  user.Roles,
				"required_roles": requiredRoles,
				"request_id":     requestID,
			})

			next(w, r, ps)
		}
	}
}

// AuthorizeStaticAPIKey validates static API key for service-to-service authentication
func AuthorizeStaticAPIKey(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := r.Context()
		requestID := getRequestID(ctx)

		logger.DebugWithContext(ctx, "Starting static API key validation")

		// Get static API key from environment
		expectedAPIKey := os.Getenv("STATIC_API_KEY")
		if expectedAPIKey == "" {
			logger.ErrorWithMetadata("Static API key not configured", map[string]interface{}{
				"request_id": requestID,
				"path":       r.URL.Path,
			}, nil)

			sendErrorResponse(w, http.StatusInternalServerError, "configuration_error", "Service authentication not configured")
			return
		}

		// Extract API key from request headers
		var providedAPIKey string
		var headerUsed string

		// Option 1: X-API-Key header (most common)
		if key := r.Header.Get("X-API-Key"); key != "" {
			providedAPIKey = key
			headerUsed = "X-API-Key"
		}

		// Option 2: Authorization header with "ApiKey" scheme
		if providedAPIKey == "" {
			if authHeader := r.Header.Get("Authorization"); authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "ApiKey " {
				providedAPIKey = authHeader[7:]
				headerUsed = "Authorization (ApiKey)"
			}
		}

		// Option 3: API-Key header (alternative)
		if providedAPIKey == "" {
			if key := r.Header.Get("API-Key"); key != "" {
				providedAPIKey = key
				headerUsed = "API-Key"
			}
		}

		// Validate API key presence
		if providedAPIKey == "" {
			logger.SecurityLog(ctx, "static_api_key_missing", false, map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"path":        r.URL.Path,
				"method":      r.Method,
				"user_agent":  r.UserAgent(),
				"request_id":  requestID,
			})

			logger.WarnWithMetadata("Static API key authentication failed - no key provided", map[string]interface{}{
				"request_id": requestID,
				"path":       r.URL.Path,
			})

			sendErrorResponse(w, http.StatusUnauthorized, "api_key_required", "Static API key is required for this endpoint")
			return
		}

		// Validate API key value
		if providedAPIKey != expectedAPIKey {
			logger.SecurityLog(ctx, "static_api_key_invalid", false, map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"path":        r.URL.Path,
				"method":      r.Method,
				"user_agent":  r.UserAgent(),
				"header_used": headerUsed,
				"key_length":  len(providedAPIKey),
				"request_id":  requestID,
			})

			logger.ErrorWithMetadata("Invalid static API key attempt", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
				"path":        r.URL.Path,
				"header_used": headerUsed,
				"request_id":  requestID,
			}, nil)

			sendErrorResponse(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key provided")
			return
		}

		logger.SecurityLog(ctx, "static_api_key_success", true, map[string]interface{}{
			"remote_addr": r.RemoteAddr,
			"path":        r.URL.Path,
			"method":      r.Method,
			"header_used": headerUsed,
			"request_id":  requestID,
		})

		logger.InfoWithMetadata("Static API key authentication successful", map[string]interface{}{
			"remote_addr": r.RemoteAddr,
			"path":        r.URL.Path,
			"header_used": headerUsed,
			"request_id":  requestID,
		})

		next(w, r, ps)
	}
}

// RequireRole is a convenience function for single role requirement
func RequireRole(role string) func(httprouter.Handle) httprouter.Handle {
	return AuthorizeWithRoles(role)
}

// RequireAdminRole requires admin role
func RequireAdminRole(next httprouter.Handle) httprouter.Handle {
	return AuthorizeWithRoles("admin")(next)
}

// RequireDataProviderRole requires data provider or admin role
func RequireDataProviderRole(next httprouter.Handle) httprouter.Handle {
	return AuthorizeWithRoles("data-provider", "admin")(next)
}

// RequireConsumerRole requires consumer, data provider, or admin role
func RequireConsumerRole(next httprouter.Handle) httprouter.Handle {
	return AuthorizeWithRoles("consumer", "data-provider", "admin")(next)
}

// Helper function to get request ID from context
func getRequestID(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

// sendErrorResponse sends a standardized JSON error response
func sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   errorCode,
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}
