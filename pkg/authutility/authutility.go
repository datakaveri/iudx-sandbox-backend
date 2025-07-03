package authutility

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents authenticated user information
type User struct {
	UserName string   `json:"username"`
	UserID   string   `json:"user_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	PublicKey     *rsa.PublicKey
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	Issuer        string
	Audience      string
}

var (
	ErrMissingToken     = errors.New("authorization token is required")
	ErrInvalidToken     = errors.New("invalid authorization token")
	ErrTokenExpired     = errors.New("authorization token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrInvalidIssuer    = errors.New("invalid token issuer")
	ErrInvalidAudience  = errors.New("invalid token audience")

	jwtConfig *JWTConfig
)

// Claims represents JWT claims structure
type Claims struct {
	UserName       string                 `json:"preferred_username"`
	UserID         string                 `json:"sub"`
	Email          string                 `json:"email,omitempty"`
	Roles          []string               `json:"realm_access.roles,omitempty"`
	ResourceAccess map[string]interface{} `json:"resource_access,omitempty"`
	jwt.RegisteredClaims
}

// InitJWTConfig initializes JWT configuration with proper validation
func InitJWTConfig() error {
	publicKeyStr := os.Getenv("KEYCLOAK_PUBLIC_KEY")
	if publicKeyStr == "" {
		return errors.New("KEYCLOAK_PUBLIC_KEY environment variable is required")
	}

	// Properly format the public key
	publicKeyPEM := fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----",
		strings.ReplaceAll(publicKeyStr, " ", "\n"))

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	jwtConfig = &JWTConfig{
		PublicKey:     publicKey,
		TokenExpiry:   time.Hour * 24,     // 24 hours
		RefreshExpiry: time.Hour * 24 * 7, // 7 days
		Issuer:        os.Getenv("JWT_ISSUER"),
		Audience:      os.Getenv("JWT_AUDIENCE"),
	}

	return nil
}

// extractTokenFromHeader extracts JWT token from Authorization header
func extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Fallback to legacy "token" header for backward compatibility
		legacyToken := r.Header.Get("token")
		if legacyToken == "" {
			return "", ErrMissingToken
		}
		// Handle legacy format: "Bearer token" or just "token"
		if strings.HasPrefix(legacyToken, "Bearer ") {
			return strings.TrimPrefix(legacyToken, "Bearer "), nil
		}
		return legacyToken, nil
	}

	// Standard Authorization header format: "Bearer <token>"
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", ErrInvalidToken
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}

// validateToken validates JWT token with proper security checks
func validateToken(tokenString string) (*jwt.Token, *Claims, error) {
	if jwtConfig == nil {
		return nil, nil, errors.New("JWT configuration not initialized")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtConfig.PublicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, nil, ErrInvalidSignature
		}
		return nil, nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, nil, ErrInvalidToken
	}

	// Additional claims validation using RegisteredClaims
	if jwtConfig.Issuer != "" && claims.RegisteredClaims.Issuer != jwtConfig.Issuer {
		return nil, nil, ErrInvalidIssuer
	}

	if jwtConfig.Audience != "" {
		validAudience := false
		for _, aud := range claims.RegisteredClaims.Audience {
			if aud == jwtConfig.Audience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, nil, ErrInvalidAudience
		}
	}

	return token, claims, nil
}

// TokenValid validates if the token in request is valid
func TokenValid(r *http.Request) error {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return err
	}

	_, _, err = validateToken(tokenString)
	return err
}

// ExtractTokenMetadata extracts user information from JWT token
func ExtractTokenMetadata(r *http.Request) (*User, error) {
	tokenString, err := extractTokenFromHeader(r)
	if err != nil {
		return nil, err
	}

	_, claims, err := validateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Extract roles from realm_access or resource_access
	var roles []string
	if claims.Roles != nil {
		roles = claims.Roles
	}

	user := &User{
		UserName: claims.UserName,
		UserID:   claims.UserID,
		Roles:    roles,
	}

	if user.UserName == "" {
		return nil, errors.New("username not found in token claims")
	}

	return user, nil
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (u *User) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}
