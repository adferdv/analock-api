package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adfer-dev/analock-api/auth"
	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/services"
	"github.com/adfer-dev/analock-api/utils"
	"github.com/gorilla/mux"
)

func InitAuthRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/auth/authenticate", utils.ParseToHandlerFunc(handleAuthenticateUser)).Methods("POST")
	router.HandleFunc("/api/v1/auth/refreshToken", utils.ParseToHandlerFunc(handleRefreshToken)).Methods("POST")
}

var authService *services.AuthService = services.NewAuthService(
	services.NewGoogleTokenValidatorImpl(),
	auth.GetTokenManager(),
	&services.UserServiceImpl{},
	&services.TokenServiceImpl{},
	&services.ExternalLoginServiceImpl{},
)

// @Summary		Authenticate user
// @Description	Authenticates a user and returns access and refresh tokens
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body		services.UserAuthenticateBody	true	"Authentication request"
// @Success		200		{object}	services.TokenResponse
// @Failure		400		{object}	models.HttpError
// @Failure		500		{object}	models.HttpError
// @Router			/auth/authenticate [post]
func handleAuthenticateUser(res http.ResponseWriter, req *http.Request) error {
	authenticateBody := services.UserAuthenticateBody{}

	validationErrs := utils.HandleValidation(req, &authenticateBody)

	if len(validationErrs) > 0 {
		return utils.WriteJSON(res, 400, validationErrs)
	}

	accessToken, refreshToken, authErr := authService.AuthenticateUser(authenticateBody)

	if authErr != nil {
		return utils.WriteJSON(res, 500, models.HttpError{Status: http.StatusInternalServerError, Description: "Error happenned when authenticating user. Please, try again."})
	}

	claims, claimsErr := authService.AppTokenManager.GetClaims(refreshToken.TokenValue)

	if claimsErr != nil {
		return claimsErr
	}
	res.Header().Add("Set-Cookie", fmt.Sprintf("refreshToken=%s; Expires=%d; HttpOnly", refreshToken.TokenValue, int64(claims["exp"].(float64))))
	return utils.WriteJSON(res, 200,
		services.TokenResponse{AccessToken: accessToken.TokenValue, RefreshToken: refreshToken.TokenValue})
}

// @Summary		Refresh access token
// @Description	Refreshes the access token using a refresh token
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			body	body		services.RefreshTokenRequest	true	"Refresh token request"
// @Success		200		{object}	services.TokenResponse
// @Failure		403		{object}	models.HttpError
// @Router			/auth/refreshToken [post]
func handleRefreshToken(res http.ResponseWriter, req *http.Request) error {
	authenticateBody := services.RefreshTokenRequest{}

	validationErrs := utils.HandleValidation(req, &authenticateBody)

	if len(validationErrs) > 0 {
		return utils.WriteJSON(res, 403, validationErrs)
	}

	newAccessToken, refreshTokenErr := authService.RefreshToken(authenticateBody)

	log.Println(refreshTokenErr)

	if refreshTokenErr != nil {
		return utils.WriteJSON(res, 403, refreshTokenErr)
	}

	return utils.WriteJSON(res, 200, newAccessToken)
}
