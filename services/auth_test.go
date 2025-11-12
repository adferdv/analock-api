package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/models"
	jwt "github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

// --- Mock Implementations for Interfaces ---

// Mock implementation for GoogleTokenValidator
type mockGoogleTokenValidator struct {
	ValidateFunc func(idToken string) error
}

func (m *mockGoogleTokenValidator) Validate(idToken string) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(idToken)
	}
	return nil
}

// Mock implementation for TokenManager
type mockTokenManager struct {
	GenerateTokenFunc func(user models.User, kind models.TokenKind) (string, error)
	ValidateTokenFunc func(tokenString string) error
	GetClaimsFunc     func(tokenString string) (jwt.MapClaims, error)
}

func (m *mockTokenManager) GenerateToken(user models.User, kind models.TokenKind) (string, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(user, kind)
	}
	if kind == models.Access {
		return constants.TestAccessTokenValue, nil
	}
	return constants.TestRefreshTokenValue, nil
}

func (m *mockTokenManager) ValidateToken(tokenString string) error {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(tokenString)
	}
	if tokenString == "valid_refresh_token" {
		return nil
	}
	return errors.New("invalid token from mock manager")
}

func (m *mockTokenManager) GetClaims(tokenString string) (jwt.MapClaims, error) {
	if m.GetClaimsFunc != nil {
		return m.GetClaimsFunc(tokenString)
	}
	if tokenString == "valid_refresh_token" {
		claims := make(jwt.MapClaims)
		claims["email"] = "exists@example.com"
		claims["exp"] = float64(time.Now().Add(1 * time.Hour).Unix())
		return claims, nil
	}
	return nil, errors.New("cannot get claims from mock manager")
}

// Mock implementation for UserService
type mockUserService struct {
	GetUserByIdFunc    func(id uint) (*models.User, error)
	GetUserByEmailFunc func(email string) (*models.User, error)
	SaveUserFunc       func(userBody UserBody) (*models.User, error)
	UpdateUserFunc     func(userBody UserBody) (*models.User, error)
	DeleteUserFunc     func(id uint) error
}

func (m *mockUserService) GetUserById(id uint) (*models.User, error) {
	if m.GetUserByIdFunc != nil {
		return m.GetUserByIdFunc(id)
	}
	if id == 1 {
		return &models.User{Id: id, Email: "byid@example.com", UserName: "User By ID"}, nil
	}
	return nil, errors.New("user not found by ID from mock service")
}

func (m *mockUserService) GetUserByEmail(email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	if email == "exists@example.com" {
		return &models.User{Id: 1, Email: email, UserName: "Existing User"}, nil
	}
	return nil, errors.New("user not found by email from mock service")
}

func (m *mockUserService) SaveUser(userBody UserBody) (*models.User, error) {
	if m.SaveUserFunc != nil {
		return m.SaveUserFunc(userBody)
	}
	return &models.User{Id: 2, Email: userBody.Email, UserName: userBody.UserName, Role: models.Standard}, nil
}

func (m *mockUserService) UpdateUser(userBody UserBody) (*models.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(userBody)
	}
	return &models.User{Email: userBody.Email, UserName: userBody.UserName, Role: models.Standard}, nil
}

func (m *mockUserService) DeleteUser(id uint) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(id)
	}
	return nil
}

// Mock implementation for TokenService
type mockTokenService struct {
	GetTokenByIdFunc       func(id uint) (*models.Token, error)
	GetTokenByValueFunc    func(tokenValue string) (*models.Token, error)
	GetUserTokenByKindFunc func(userId uint, kind models.TokenKind) (*models.Token, error)
	UpdateTokenFunc        func(tokenBody *models.Token) (*models.Token, error)
	SaveTokenFunc          func(tokenBody *models.Token) (*models.Token, error)
	GetUserTokenPairFunc   func(userId uint) ([2]*models.Token, error)
	DeleteTokenFunc        func(id uint) error
}

func (m *mockTokenService) GetTokenById(id uint) (*models.Token, error) {
	if m.GetTokenByIdFunc != nil {
		return m.GetTokenByIdFunc(id)
	}
	return &models.Token{Id: id, TokenValue: "mock_token_by_id"}, nil
}

func (m *mockTokenService) GetTokenByValue(tokenValue string) (*models.Token, error) {
	if m.GetTokenByValueFunc != nil {
		return m.GetTokenByValueFunc(tokenValue)
	}
	return &models.Token{TokenValue: tokenValue, Id: 99}, nil
}

func (m *mockTokenService) GetUserTokenByKind(userId uint, kind models.TokenKind) (*models.Token, error) {
	if m.GetUserTokenByKindFunc != nil {
		return m.GetUserTokenByKindFunc(userId, kind)
	}
	if userId == 1 && kind == models.Access {
		return &models.Token{Id: 1, TokenValue: "db_access_token", Kind: models.Access, UserRefer: userId}, nil
	}
	return nil, errors.New("token not found by mock service")
}

func (m *mockTokenService) UpdateToken(tokenBody *models.Token) (*models.Token, error) {
	if m.UpdateTokenFunc != nil {
		return m.UpdateTokenFunc(tokenBody)
	}
	return tokenBody, nil
}

func (m *mockTokenService) SaveToken(tokenBody *models.Token) (*models.Token, error) {
	if m.SaveTokenFunc != nil {
		return m.SaveTokenFunc(tokenBody)
	}
	return tokenBody, nil
}

func (m *mockTokenService) GetUserTokenPair(userId uint) ([2]*models.Token, error) {
	if m.GetUserTokenPairFunc != nil {
		return m.GetUserTokenPairFunc(userId)
	}
	return [2]*models.Token{
		{Id: 1, TokenValue: "pair_access_token", Kind: models.Access, UserRefer: userId},
		{Id: 2, TokenValue: "pair_refresh_token", Kind: models.Refresh, UserRefer: userId},
	}, nil
}

func (m *mockTokenService) DeleteToken(id uint) error {
	if m.DeleteTokenFunc != nil {
		return m.DeleteTokenFunc(id)
	}
	return nil
}

// Mock implementation for ExternalLoginService
type mockExternalLoginService struct {
	GetExternalLoginByIdFunc         func(id uint) (*models.ExternalLogin, error)
	GetExternalLoginByClientIdFunc   func(clientId string) (*models.ExternalLogin, error)
	SaveExternalLoginFunc            func(externalLoginBody *models.ExternalLogin) (*models.ExternalLogin, error)
	UpdateExternalLoginFunc          func(externalLoginBody *models.ExternalLogin) (*models.ExternalLogin, error)
	UpdateUserExternalLoginTokenFunc func(userId uint, externalLoginBody *UpdateExternalLoginBody) (*models.ExternalLogin, error)
	DeleteExternalLoginFunc          func(id uint) error
}

func (m *mockExternalLoginService) GetExternalLoginById(id uint) (*models.ExternalLogin, error) {
	if m.GetExternalLoginByIdFunc != nil {
		return m.GetExternalLoginByIdFunc(id)
	}
	return &models.ExternalLogin{Id: id, ClientId: "client_id_by_id"}, nil
}

func (m *mockExternalLoginService) GetExternalLoginByClientId(clientId string) (*models.ExternalLogin, error) {
	if m.GetExternalLoginByClientIdFunc != nil {
		return m.GetExternalLoginByClientIdFunc(clientId)
	}
	return &models.ExternalLogin{ClientId: clientId, Id: 99}, nil
}

func (m *mockExternalLoginService) SaveExternalLogin(externalLoginBody *models.ExternalLogin) (*models.ExternalLogin, error) {
	if m.SaveExternalLoginFunc != nil {
		return m.SaveExternalLoginFunc(externalLoginBody)
	}
	return externalLoginBody, nil
}

func (m *mockExternalLoginService) UpdateExternalLogin(externalLoginBody *models.ExternalLogin) (*models.ExternalLogin, error) {
	if m.UpdateExternalLoginFunc != nil {
		return m.UpdateExternalLoginFunc(externalLoginBody)
	}
	return externalLoginBody, nil
}

func (m *mockExternalLoginService) UpdateUserExternalLoginToken(userId uint, externalLoginBody *UpdateExternalLoginBody) (*models.ExternalLogin, error) {
	if m.UpdateUserExternalLoginTokenFunc != nil {
		return m.UpdateUserExternalLoginTokenFunc(userId, externalLoginBody)
	}
	return &models.ExternalLogin{Id: 1, UserRefer: userId, ClientToken: externalLoginBody.ClientToken, Provider: models.Google}, nil
}

func (m *mockExternalLoginService) DeleteExternalLogin(id uint) error {
	if m.DeleteExternalLoginFunc != nil {
		return m.DeleteExternalLoginFunc(id)
	}
	return nil
}

// -- Test functions --

// Mock HTTP server for Google token validation (can still be used by mockGoogleTokenValidator)
var mockGoogleServer *httptest.Server

func TestAuthenticateUser_ExistingUser(t *testing.T) {
	mockGoogleServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGoogleServer.Close()

	googleVal := NewGoogleTokenValidatorImpl()
	googleVal.Client = mockGoogleServer.Client()
	googleVal.TokenInfoBaseURL = mockGoogleServer.URL

	mockAppTokenMgr := &mockTokenManager{}
	mockUserSvc := &mockUserService{}
	mockTokenSvc := &mockTokenService{}
	mockExtLoginSvc := &mockExternalLoginService{}

	authService := NewAuthService(googleVal, mockAppTokenMgr, mockUserSvc, mockTokenSvc, mockExtLoginSvc)

	authBody := UserAuthenticateBody{
		Email:         "exists@example.com",
		UserName:      "Existing User",
		ProviderId:    "google123",
		ProviderToken: "valid_google_token",
	}

	accessToken, refreshToken, err := authService.AuthenticateUser(authBody)

	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotNil(t, refreshToken)
	assert.Equal(t, constants.TestAccessTokenValue, accessToken.TokenValue)
	assert.Equal(t, constants.TestRefreshTokenValue, refreshToken.TokenValue)
}

func TestAuthenticateUser_NewUser(t *testing.T) {
	mockGoogleServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockGoogleServer.Close()

	mockGoogleVal := &mockGoogleTokenValidator{
		ValidateFunc: func(idToken string) error { return nil }, // Assume valid
	}
	mockAppTokenMgr := &mockTokenManager{}
	mockUserSvc := &mockUserService{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return nil, errors.New("user not found for new user test") // Force new user path
		},
	}
	mockTokenSvc := &mockTokenService{}
	mockExtLoginSvc := &mockExternalLoginService{}

	authService := NewAuthService(mockGoogleVal, mockAppTokenMgr, mockUserSvc, mockTokenSvc, mockExtLoginSvc)

	authBody := UserAuthenticateBody{
		Email:         "new@example.com",
		UserName:      "New User",
		ProviderId:    "google456",
		ProviderToken: "valid_google_token",
	}

	accessToken, refreshToken, err := authService.AuthenticateUser(authBody)

	assert.NoError(t, err)
	assert.NotNil(t, accessToken)
	assert.NotNil(t, refreshToken)
	assert.Equal(t, constants.TestAccessTokenValue, accessToken.TokenValue)
	assert.Equal(t, constants.TestRefreshTokenValue, refreshToken.TokenValue)
}

func TestAuthenticateUser_GoogleTokenInvalid(t *testing.T) {
	mockGoogleServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer mockGoogleServer.Close()

	googleVal := NewGoogleTokenValidatorImpl()
	googleVal.Client = mockGoogleServer.Client()
	googleVal.TokenInfoBaseURL = mockGoogleServer.URL

	mockAppTokenMgr := &mockTokenManager{}
	mockUserSvc := &mockUserService{}
	mockTokenSvc := &mockTokenService{}
	mockExtLoginSvc := &mockExternalLoginService{}

	authService := NewAuthService(googleVal, mockAppTokenMgr, mockUserSvc, mockTokenSvc, mockExtLoginSvc)

	authBody := UserAuthenticateBody{
		Email:         "test@example.com",
		UserName:      "Test User",
		ProviderId:    "google789",
		ProviderToken: "invalid_google_token",
	}

	_, _, err := authService.AuthenticateUser(authBody)

	assert.Error(t, err)
	assert.EqualError(t, err, "google token not valid")
}

func TestRefreshToken_Valid(t *testing.T) {
	mockTokenManager := &mockTokenManager{
		ValidateTokenFunc: func(tokenString string) error { return nil },
		GetClaimsFunc: func(tokenString string) (jwt.MapClaims, error) {
			return jwt.MapClaims{"sub": float64(1)}, nil
		},
	}
	mockUserService := &mockUserService{
		GetUserByEmailFunc: func(email string) (*models.User, error) {
			return &models.User{Id: 1, Email: email, UserName: "Test User"}, nil
		},
	}
	mockTokenService := &mockTokenService{}

	authService := NewAuthService(nil, mockTokenManager, mockUserService, mockTokenService, nil)

	req := RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	res, err := authService.RefreshToken(req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, constants.TestAccessTokenValue, res.Token)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockAppTokenMgr := &mockTokenManager{
		ValidateTokenFunc: func(tokenString string) error {
			return errors.New("invalid token from test")
		},
	}
	authService := NewAuthService(nil, mockAppTokenMgr, nil, nil, nil)

	req := RefreshTokenRequest{
		RefreshToken: "invalid_token_for_refresh",
	}

	res, err := authService.RefreshToken(req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, "invalid token from test")
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	mockAppTokenMgr := &mockTokenManager{
		ValidateTokenFunc: func(tokenString string) error { return nil },
		GetClaimsFunc: func(tokenString string) (jwt.MapClaims, error) {
			return jwt.MapClaims{"sub": float64(1)}, nil
		},
	}
	mockUserSvc := &mockUserService{
		GetUserByIdFunc: func(userId uint) (*models.User, error) {
			if userId == 1 {
				return nil, errors.New("user not found for refresh test")
			}
			return nil, errors.New("unexpected email in GetUserByEmail mock")
		},
	}
	mockTokenService := &mockTokenService{}
	authService := NewAuthService(nil, mockAppTokenMgr, mockUserSvc, mockTokenService, nil)

	req := RefreshTokenRequest{
		RefreshToken: "valid_refresh_token_unknown_user",
	}

	res, err := authService.RefreshToken(req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, "user not found for refresh test")
}
