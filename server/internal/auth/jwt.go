package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateTokenPair(userID string, secret []byte, accessTTL time.Duration, refreshTTL time.Duration) (TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(accessTTL)
	refreshExp := now.Add(refreshTTL)

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString(secret)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString(secret)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func ParseToken(token string, secret []byte) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
