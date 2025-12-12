package utils

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/adfer-dev/analock-api/models"
	"github.com/go-playground/validator/v10"
)

// Writes the given value structure as an HTTP response with the given status.
func WriteJSON(res http.ResponseWriter, status int, value any) error {
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(status)

	return json.NewEncoder(res).Encode(value)
}

// Parses JSON from reader and fits it into the given body structure
func ReadJSON(reader io.Reader, body interface{}) error {
	if deserializeErr := json.NewDecoder(reader).Decode(body); deserializeErr != nil {
		return deserializeErr
	}

	if validationErr := validateBody(body); validationErr != nil {
		return validationErr
	}

	return nil
}

// Wrapper that writes an error as an HTTP response with given info.
func WriteError(res http.ResponseWriter, status int, description string) error {
	return WriteJSON(
		res,
		status,
		&models.HttpError{
			Status:      status,
			Description: description,
		},
	)
}

func validateBody(body interface{}) error {
	newValidator := validator.New()

	if err := newValidator.Struct(body); err != nil {
		return err
	}

	return nil
}
