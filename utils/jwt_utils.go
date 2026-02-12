package utils

import (
	"errors"
	"os"
	"time"

	constants "lem-be/constants"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines the structure of JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   constants.Role `json:"role"`
	jwt.RegisteredClaims
}

// GetJWTSecret retrieves the JWT secret from environment variables
func GetJWTSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		NewLogger("JWTUtils", "GetJWTSecret").Errorf("JWT_SECRET environment variable is not set")
		return "", errors.New("JWT_SECRET environment variable is not set")
	}
	return secret, nil
}

// GenerateAccessToken generates a short-lived access token (15 minutes)
func GenerateAccessToken(userID, email string, role constants.Role) (string, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generates a long-lived refresh token (7 days)
func GenerateRefreshToken(userID string) (string, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		NewLogger("JWTUtils", "ValidateToken").Errorf("Failed to get JWT secret: %v", err)
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		NewLogger("JWTUtils", "ValidateToken").Warnf("Token parsing failed: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

