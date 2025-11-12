package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/adfer-dev/analock-api/models"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

var testSecretKey = []byte("test-secret-key-for-token-manager")

// mockSecretKeyProvider returns a fixed secret key for testing.
func mockSecretKeyProvider() ([]byte, error) {
	return testSecretKey, nil
}

// mockErrorSecretKeyProvider returns an error, simulating failure to get the key.
func mockErrorSecretKeyProvider() ([]byte, error) {
	return nil, errors.New("mock secret key provider error")
}

func TestDefaultTokenManager_GenerateToken(t *testing.T) {
	manager := newDefaultTokenManagerWithProvider(mockSecretKeyProvider)
	user := models.User{Id: 1, Email: "test@example.com"}

	t.Run("generate_access_token", func(t *testing.T) {
		tokenString, err := manager.GenerateToken(user, models.Access)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return testSecretKey, nil
		})
		assert.NoError(t, parseErr)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, user.Id, uint(claims["sub"].(float64)))
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		assert.InDelta(t, time.Now().Add(1*time.Hour).Unix(), int64(exp), 5)
	})

	t.Run("generate_refresh_token", func(t *testing.T) {
		tokenString, err := manager.GenerateToken(user, models.Refresh)
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		token, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return testSecretKey, nil
		})
		assert.NoError(t, parseErr)
		assert.True(t, token.Valid)

		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, user.Id, uint(claims["sub"].(float64)))
		exp, ok := claims["exp"].(float64)
		assert.True(t, ok)
		assert.InDelta(t, time.Now().Add(24*7*time.Hour).Unix(), int64(exp), 5)
	})

	t.Run("error_from_get_secret_key", func(t *testing.T) {
		errorManager := newDefaultTokenManagerWithProvider(mockErrorSecretKeyProvider)
		_, err := errorManager.GenerateToken(user, models.Access)
		assert.Error(t, err)
		assert.EqualError(t, err, "mock secret key provider error")
	})
}

func TestDefaultTokenManager_ValidateToken(t *testing.T) {
	manager := newDefaultTokenManagerWithProvider(mockSecretKeyProvider)
	user := models.User{Id: 1, Email: "test@example.com"}
	validAccessToken, _ := manager.GenerateToken(user, models.Access)

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	})
	expiredTokenString, _ := expiredToken.SignedString(testSecretKey)

	otherSecret := []byte("other-secret-key")
	tokenWithOtherKey := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenWithOtherKeyString, _ := tokenWithOtherKey.SignedString(otherSecret)

	t.Run("validate_valid_token", func(t *testing.T) {
		err := manager.ValidateToken(validAccessToken)
		assert.NoError(t, err)
	})

	t.Run("validate_expired_token", func(t *testing.T) {
		err := manager.ValidateToken(expiredTokenString)
		assert.Error(t, err)
		ve, ok := err.(*jwt.ValidationError)
		assert.True(t, ok, "error should be a *jwt.ValidationError")
		if ok {
			assert.True(t, (ve.Errors&jwt.ValidationErrorExpired != 0) || (ve.Errors&jwt.ValidationErrorNotValidYet != 0), "ValidationError should be due to token being expired or not yet valid")
		}
	})

	t.Run("validate_token_wrong_key", func(t *testing.T) {
		err := manager.ValidateToken(tokenWithOtherKeyString)
		assert.Error(t, err)
		ve, ok := err.(*jwt.ValidationError)
		assert.True(t, ok, "error should be a *jwt.ValidationError")
		if ok {
			assert.True(t, ve.Errors&jwt.ValidationErrorSignatureInvalid != 0, "ValidationError should be due to invalid signature")
		}
	})

	t.Run("validate_token_invalid_signing_method", func(t *testing.T) {
		noneAlgToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ."
		err := manager.ValidateToken(noneAlgToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signing method not valid")

		errMalformed := manager.ValidateToken("this.is.not.a.jwt")
		assert.Error(t, errMalformed)
		ve, ok := errMalformed.(*jwt.ValidationError)
		assert.True(t, ok, "error should be a *jwt.ValidationError")
		if ok {
			assert.True(t, ve.Errors&jwt.ValidationErrorMalformed != 0, "ValidationError should be due to malformed token")
		}
	})

	t.Run("error_from_get_secret_key_on_validate", func(t *testing.T) {
		errorManager := newDefaultTokenManagerWithProvider(mockErrorSecretKeyProvider)
		err := errorManager.ValidateToken(validAccessToken)
		assert.Error(t, err)
		assert.EqualError(t, err, "mock secret key provider error")
	})
}

func TestDefaultTokenManager_GetClaims(t *testing.T) {
	manager := newDefaultTokenManagerWithProvider(mockSecretKeyProvider)
	user := models.User{Id: 1, Email: "test@example.com"}
	validAccessToken, _ := manager.GenerateToken(user, models.Access)

	t.Run("get_claims_valid_token", func(t *testing.T) {
		claims, err := manager.GetClaims(validAccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, user.Id, uint(claims["sub"].(float64)))
		_, ok := claims["exp"].(float64)
		assert.True(t, ok)
	})

	t.Run("get_claims_invalid_token_signature", func(t *testing.T) {
		otherSecret := []byte("other-secret-for-claims")
		tokenWithOtherKey := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.Id,
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenWithOtherKeyString, _ := tokenWithOtherKey.SignedString(otherSecret)

		_, err := manager.GetClaims(tokenWithOtherKeyString)
		assert.Error(t, err)
		ve, ok := err.(*jwt.ValidationError)
		assert.True(t, ok, "error should be a *jwt.ValidationError")
		if ok {
			assert.True(t, ve.Errors&jwt.ValidationErrorSignatureInvalid != 0, "ValidationError should be due to invalid signature")
		}
	})

	t.Run("get_claims_malformed_token", func(t *testing.T) {
		_, err := manager.GetClaims("this.is.not.a.jwt")
		assert.Error(t, err)
		ve, ok := err.(*jwt.ValidationError)
		assert.True(t, ok, "error should be a *jwt.ValidationError")
		if ok {
			assert.True(t, ve.Errors&jwt.ValidationErrorMalformed != 0, "ValidationError should be due to malformed token")
		}
	})

	t.Run("error_from_get_secret_key_on_get_claims", func(t *testing.T) {
		errorManager := newDefaultTokenManagerWithProvider(mockErrorSecretKeyProvider)
		_, err := errorManager.GetClaims(validAccessToken)
		assert.Error(t, err)
		assert.EqualError(t, err, "mock secret key provider error")
	})
}
