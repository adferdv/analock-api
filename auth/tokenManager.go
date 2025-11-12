package auth

import (
	"errors"
	"time"

	"github.com/adfer-dev/analock-api/models"
	"github.com/golang-jwt/jwt"
)

// TokenManager interface
type TokenManager interface {
	GenerateToken(user models.User, kind models.TokenKind) (string, error)
	ValidateToken(tokenString string) error
	GetClaims(tokenString string) (jwt.MapClaims, error)
}

// Interface implementation for TokenManager
type TokenManagerImpl struct {
	secretKeyProvider func() ([]byte, error)
}

var tokenManagerInstance *TokenManagerImpl

func GetTokenManager() *TokenManagerImpl {
	if tokenManagerInstance == nil {
		tokenManagerInstance = newTokenManagerImpl()
	}

	return tokenManagerInstance
}

// Empty constructor for TokenManager implementation. Uses default secret key provider.
func newTokenManagerImpl() *TokenManagerImpl {
	return &TokenManagerImpl{secretKeyProvider: GetSecretKey}
}

// Parametrized constructor for DefaultTokenManager
// If the provided provider is nil, it defaults to using default secret key provider.
func newDefaultTokenManagerWithProvider(provider func() ([]byte, error)) *TokenManagerImpl {
	p := provider
	if p == nil {
		p = GetSecretKey
	}
	return &TokenManagerImpl{secretKeyProvider: p}
}

func (d *TokenManagerImpl) GenerateToken(user models.User, kind models.TokenKind) (string, error) {
	secretKey, envErr := d.secretKeyProvider()

	if envErr != nil {
		return "", envErr
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	var expiration int64

	if kind == models.Access {
		expiration = time.Now().Add(1 * time.Hour).Unix()
	} else {
		expiration = time.Now().Add(24 * 7 * time.Hour).Unix()
	}

	claims["sub"] = user.Id
	claims["exp"] = expiration

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (d *TokenManagerImpl) ValidateToken(tokenString string) error {
	secretKey, envErr := d.secretKeyProvider()

	if envErr != nil {
		return envErr
	}

	token, parseErr := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, errors.New("signing method not valid")
		}

		return secretKey, nil
	})

	if token == nil && parseErr != nil {
		return parseErr
	}

	if parseErr != nil || !token.Valid {
		return parseErr
	}

	return nil
}

func (d *TokenManagerImpl) GetClaims(tokenString string) (jwt.MapClaims, error) {
	secretKey, envErr := d.secretKeyProvider()

	if envErr != nil {
		return nil, envErr
	}

	jwtToken, parseErr := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if parseErr != nil {
		return nil, parseErr
	}

	if jwtToken == nil {
		return nil, errors.New("failed to parse token, token is nil")
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not assert token claims to jwt.MapClaims")
	}

	return claims, nil
}
