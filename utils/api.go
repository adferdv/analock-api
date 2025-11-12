package utils

import (
	"fmt"
	"net/http"

	"github.com/adfer-dev/analock-api/auth"
	"github.com/adfer-dev/analock-api/models"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
)

type APIFunc func(res http.ResponseWriter, req *http.Request) error

var tokenManager *auth.TokenManagerImpl = auth.GetTokenManager()

// Function that parses an APIFunc function to a http.HandlerFunc function
func ParseToHandlerFunc(f APIFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		if err := f(res, req); err != nil {
			WriteJSON(res, 500, err.Error())
		}

	}
}

// Maps an error to the HttpError struct
func TranslateDbErrorToHttpError(err error) *models.HttpError {
	httpError := &models.HttpError{}

	switch err.(type) {
	case *models.DbNotFoundError:
		httpError.Status = 404
	case *models.DbCouldNotParseItemError:
		httpError.Status = 500
	case *models.DbItemAlreadyExistsError:
		httpError.Status = 400
	default:
		httpError.Status = 500
	}

	httpError.Description = err.Error()

	return httpError
}

// Handles request body validation, returning validation errors if found.
func HandleValidation(req *http.Request, body interface{}) []*models.HttpError {
	httpErrors := make([]*models.HttpError, 0)

	if parseErr := ReadJSON(req.Body, body); parseErr != nil {
		GetCustomLogger().Info(parseErr)
		if validationErrs, ok := parseErr.(validator.ValidationErrors); ok {
			for _, validationErr := range validationErrs {
				httpErrors = append(httpErrors,
					&models.HttpError{Status: 400, Description: "Field" + validationErr.Field() + " must be provided."})
			}
		} else {
			httpError := models.HttpError{Status: 400, Description: "Not valid JSON."}
			httpErrors = append(httpErrors, &httpError)
		}
	}

	return httpErrors
}

// Builds a cache key based on given user ID
func BuildUserCacheKey(userId uint) string {
	return fmt.Sprintf("user-%d", userId)
}

// Builds a cache key based on given user ID, start date and end date
func BuildUserDateRangeCacheKey(userId uint, startDate int, endDate int) string {
	return fmt.Sprintf("%s-start%d-end%d", BuildUserCacheKey(userId), startDate, endDate)
}

// Gets token claims, by first retrieving token value from HTTP headers
func GetTokenClaimsFromRequest(req *http.Request) (jwt.MapClaims, error) {
	tokenValue := req.Header.Get("Authorization")[7:]
	tokenClaims, claimsErr := tokenManager.GetClaims(tokenValue)

	if claimsErr != nil {
		return nil, claimsErr
	}

	return tokenClaims, nil
}
