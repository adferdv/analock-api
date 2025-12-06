package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/adfer-dev/analock-api/utils"
)

// Performs an HTTP request with given method, URL, body and retry count.
func PerformRequest[T any](method string, url string, body interface{}) (*T, error) {
	utils.GetCustomLogger().Infof(
		"HTTP request: [%s]%s\n",
		method,
		url,
	)
	var bodyReader io.Reader = nil

	if body != nil {
		bodyBytes, marshallErr := json.Marshal(body)

		if marshallErr != nil {
			return nil, marshallErr
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}
	request, buildReqErr := http.NewRequest(method, url, bodyReader)
	request.Header.Set("Content-Type", "application/json")

	if buildReqErr != nil {
		return nil, buildReqErr
	}

	client := utils.GetDefaultHttpClient()

	// wrap request execution inside a function variable
	reqExecution := func() (io.ReadCloser, error) {
		response, reqErr := client.Do(request)

		if reqErr != nil {
			return nil, reqErr
		}

		// if response status is 4xx or 5xx, return error so it is retried
		if response.StatusCode >= 200 {
			utils.GetCustomLogger().Errorf(
				"Error on HTTP request: [%s]%s - STATUS: %d\n",
				method,
				url,
				response.StatusCode,
			)
			return nil, errors.New("request error")
		}
		return response.Body, nil
	}

	responseBody, reqErr := retry(reqExecution, 5)

	if reqErr != nil {
		return nil, reqErr
	}
	defer responseBody.Close()

	// Write response body into given data structure
	var res T
	readErr := utils.ReadJSON(responseBody, &res)

	if readErr != nil {
		utils.GetCustomLogger().Errorf(
			"Error on response unmarshal: %s\n",
			readErr.Error(),
		)
		return nil, readErr
	}

	return &res, nil
}

// Executes the HTTP request that is wrapped in given function and retries it.
//
// It retries for the given maximum number of retries.
// The retry interval is exponential, being doubled for each retry (starting at 1s).
func retry(f func() (io.ReadCloser, error), maxRetries uint) (io.ReadCloser, error) {
	// First execute request
	if res, err := f(); err == nil {
		return res, nil
	}

	// If request fails, retry
	var currentRetries uint = 0
	interval := 1000

	for currentRetries < maxRetries {
		utils.GetCustomLogger().Errorf(
			"Performing request retry... %d retries left.\n",
			maxRetries-currentRetries,
		)
		if res, err := f(); err == nil {
			return res, nil
		}
		currentRetries++
		interval *= 2
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	return nil, errors.New("request error")
}
