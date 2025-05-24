package pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtUtil struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessExpiry       time.Duration
	RefreshExpiry      time.Duration
}

type JwtClaim struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJwtUtil(accessSecret, refreshSecret string, accessExpiry, refreshExpiry time.Duration) *JwtUtil {
	return &JwtUtil{
		AccessTokenSecret:  accessSecret,
		RefreshTokenSecret: refreshSecret,
		AccessExpiry:       accessExpiry,
		RefreshExpiry:      refreshExpiry,
	}
}

// GenerateAccessToken creates a new access token for the given user ID
func (j *JwtUtil) GenerateAccessToken(userID string) (string, error) {
	claims := &JwtClaim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.AccessTokenSecret))
}

// GenerateRefreshToken creates a new refresh token for the given user ID
func (j *JwtUtil) GenerateRefreshToken(userID string) (string, error) {
	claims := &JwtClaim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.RefreshTokenSecret))
}

// ValidateAccessToken validates the access token and returns the claims
func (j *JwtUtil) ValidateAccessToken(tokenString string) (*JwtClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaim{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtClaim); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates the refresh token and returns the claims
func (j *JwtUtil) ValidateRefreshToken(tokenString string) (*JwtClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaim{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtClaim); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}
