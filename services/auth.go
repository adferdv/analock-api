package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/adfer-dev/analock-api/auth"
	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/models"
)

// AuthService struct
type AuthService struct {
	googleValidator GoogleTokenValidator
	AppTokenManager auth.TokenManager
	userService     UserService
	tokenService    TokenService
	extLoginService ExternalLoginService
}

// AuthService constructor
func NewAuthService(
	googleValidator GoogleTokenValidator,
	appTokenManager auth.TokenManager,
	userService UserService,
	tokenService TokenService,
	extLoginService ExternalLoginService,
) *AuthService {
	return &AuthService{
		googleValidator: googleValidator,
		AppTokenManager: appTokenManager,
		userService:     userService,
		tokenService:    tokenService,
		extLoginService: extLoginService,
	}
}

// Request bodies
type UserAuthenticateBody struct {
	Email         string `json:"email" validate:"required,email"`
	UserName      string `json:"userName" validate:"required"`
	ProviderId    string `json:"providerId" validate:"required"`
	ProviderToken string `json:"providerToken" validate:"required,jwt"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}

// AuthService methods
func (authService *AuthService) AuthenticateUser(authBody UserAuthenticateBody) (*models.Token, *models.Token, error) {
	googleValidateErr := authService.validateGoogleToken(authBody.ProviderToken)
	if googleValidateErr != nil {
		return nil, nil, googleValidateErr
	}

	user, getUserErr := authService.userService.GetUserByEmail(authBody.Email)

	if getUserErr == nil {
		externalLogin := &UpdateExternalLoginBody{
			ClientToken: authBody.ProviderToken,
		}
		_, saveExternalLoginError := authService.extLoginService.UpdateUserExternalLoginToken(user.Id, externalLogin)
		if saveExternalLoginError != nil {
			return nil, nil, saveExternalLoginError
		}
		return authService.updateTokenPair(user)
	} else {
		userBody := UserBody{
			Email:    authBody.Email,
			UserName: authBody.UserName,
		}
		savedUser, saveUserError := authService.userService.SaveUser(userBody)
		if saveUserError != nil {
			return nil, nil, saveUserError
		}

		externalLogin := &models.ExternalLogin{
			ClientId:    authBody.ProviderId,
			ClientToken: authBody.ProviderToken,
			UserRefer:   savedUser.Id,
			Provider:    models.Google,
		}
		_, saveExternalLoginError := authService.extLoginService.SaveExternalLogin(externalLogin)
		if saveExternalLoginError != nil {
			// Consider rolling back user creation or logging, for now, return error
			return nil, nil, saveExternalLoginError
		}
		return authService.generateAndSaveTokenPair(savedUser)
	}
}

func (authService *AuthService) RefreshToken(request RefreshTokenRequest) (*RefreshTokenResponse, error) {
	validationErr := authService.AppTokenManager.ValidateToken(request.RefreshToken)
	if validationErr != nil {
		return nil, validationErr
	}

	claims, claimsErr := authService.AppTokenManager.GetClaims(request.RefreshToken)
	if claimsErr != nil {
		return nil, claimsErr
	}

	userId, ok := claims["sub"].(float64)

	if !ok {
		return nil, errors.New("user id is not a number or not found")
	}

	user, getUserErr := authService.userService.GetUserById(uint(userId))

	if getUserErr != nil {
		return nil, getUserErr
	}

	accessTokenString, accessTokenErr := authService.AppTokenManager.GenerateToken(*user, models.Access)
	if accessTokenErr != nil {
		return nil, accessTokenErr
	}

	dbAccessToken, getDbAccessTokenErr := authService.tokenService.GetUserTokenByKind(user.Id, models.Access)
	if getDbAccessTokenErr != nil {
		return nil, getDbAccessTokenErr
	}

	accessToken := &models.Token{
		Id:         dbAccessToken.Id,
		TokenValue: accessTokenString,
		Kind:       models.Access,
		UserRefer:  user.Id,
	}

	_, saveAccessTokenErr := authService.tokenService.UpdateToken(accessToken)
	if saveAccessTokenErr != nil {
		return nil, saveAccessTokenErr
	}

	return &RefreshTokenResponse{Token: accessToken.TokenValue}, nil
}

func (authService *AuthService) generateAndSaveTokenPair(user *models.User) (accessToken *models.Token, refreshToken *models.Token, err error) {
	accessTokenString, accessTokenErr := authService.AppTokenManager.GenerateToken(*user, models.Access)
	if accessTokenErr != nil {
		return nil, nil, accessTokenErr
	}
	accessToken = &models.Token{
		TokenValue: accessTokenString,
		Kind:       models.Access,
		UserRefer:  user.Id,
	}

	refreshTokenString, refreshTokenErr := authService.AppTokenManager.GenerateToken(*user, models.Refresh)
	if refreshTokenErr != nil {
		return nil, nil, refreshTokenErr
	}
	refreshToken = &models.Token{
		TokenValue: refreshTokenString,
		Kind:       models.Refresh,
		UserRefer:  user.Id,
	}

	_, saveAccessTokenErr := authService.tokenService.SaveToken(accessToken)
	if saveAccessTokenErr != nil {
		return nil, nil, saveAccessTokenErr
	}

	_, saveRefreshTokenErr := authService.tokenService.SaveToken(refreshToken)
	if saveRefreshTokenErr != nil {
		// Consider cleanup for already saved access token
		return nil, nil, saveRefreshTokenErr
	}
	return accessToken, refreshToken, nil
}

func (authService *AuthService) updateTokenPair(user *models.User) (accessToken *models.Token, refreshToken *models.Token, err error) {
	tokenPair, getTokenPairErr := authService.tokenService.GetUserTokenPair(user.Id)
	if getTokenPairErr != nil {
		return nil, nil, getTokenPairErr
	}

	accessTokenString, accessTokenErr := authService.AppTokenManager.GenerateToken(*user, models.Access)
	if accessTokenErr != nil {
		return nil, nil, accessTokenErr
	}

	refreshTokenString, refreshTokenErr := authService.AppTokenManager.GenerateToken(*user, models.Refresh)
	if refreshTokenErr != nil {
		return nil, nil, refreshTokenErr
	}

	var updatedAccess, updatedRefresh *models.Token

	for _, token := range tokenPair {
		var currentTokenToUpdate *models.Token
		if token.Kind == models.Access {
			token.TokenValue = accessTokenString
			updatedAccess = token
			currentTokenToUpdate = updatedAccess
		} else if token.Kind == models.Refresh {
			token.TokenValue = refreshTokenString
			updatedRefresh = token
			currentTokenToUpdate = updatedRefresh
		}

		if currentTokenToUpdate != nil {
			_, updateErr := authService.tokenService.UpdateToken(currentTokenToUpdate)
			if updateErr != nil {
				return nil, nil, updateErr // return early on first error
			}
		}
	}

	if updatedAccess == nil || updatedRefresh == nil {
		return nil, nil, errors.New("failed to update token pair, one or both tokens not found in existing pair")
	}

	return updatedAccess, updatedRefresh, nil
}

func (authService *AuthService) validateGoogleToken(idToken string) error {
	return authService.googleValidator.Validate(idToken)
}

// Interfaces and implementations for the GoogleTokenValidator

// GoogleTokenValidator interface
type GoogleTokenValidator interface {
	Validate(idToken string) error
}

// Interface implementation for GoogleTokenValidator
type GoogleTokenValidatorImpl struct {
	Client           *http.Client
	TokenInfoBaseURL string
}

// Constructor for GoogleTokenValidator implementation.
// Sets TokenInfoBaseUrl to default Google token validation URL.
func NewGoogleTokenValidatorImpl() *GoogleTokenValidatorImpl {
	return &GoogleTokenValidatorImpl{
		TokenInfoBaseURL: constants.ApiGoogleTokenValidationUrl,
	}
}

// Validates the Google token
func (d *GoogleTokenValidatorImpl) Validate(idToken string) error {
	httpClient := d.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	reqURL := fmt.Sprintf("%s?id_token=%s", d.TokenInfoBaseURL, idToken)
	googleAuthRes, googleAuthReqErr := httpClient.Get(reqURL)
	if googleAuthReqErr != nil {
		return googleAuthReqErr
	}
	defer googleAuthRes.Body.Close()

	if googleAuthRes.StatusCode != http.StatusOK {
		log.Printf("Google token validation failed with status: %s", googleAuthRes.Status)
		return errors.New("google token not valid")
	}
	return nil
}
