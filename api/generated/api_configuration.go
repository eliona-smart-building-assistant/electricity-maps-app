// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Electricity Maps app API
 *
 * API to access and configure the Electricity Maps app
 *
 * API version: 1.0.0
 */

package apiserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// ConfigurationAPIController binds http requests to an api service and writes the service results to the http response
type ConfigurationAPIController struct {
	service      ConfigurationAPIServicer
	errorHandler ErrorHandler
}

// ConfigurationAPIOption for how the controller is set up.
type ConfigurationAPIOption func(*ConfigurationAPIController)

// WithConfigurationAPIErrorHandler inject ErrorHandler into controller
func WithConfigurationAPIErrorHandler(h ErrorHandler) ConfigurationAPIOption {
	return func(c *ConfigurationAPIController) {
		c.errorHandler = h
	}
}

// NewConfigurationAPIController creates a default api controller
func NewConfigurationAPIController(s ConfigurationAPIServicer, opts ...ConfigurationAPIOption) *ConfigurationAPIController {
	controller := &ConfigurationAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ConfigurationAPIController
func (c *ConfigurationAPIController) Routes() Routes {
	return Routes{
		"GetConfiguration": Route{
			strings.ToUpper("Get"),
			"/v1/configs",
			c.GetConfiguration,
		},
		"PutConfiguration": Route{
			strings.ToUpper("Put"),
			"/v1/configs",
			c.PutConfiguration,
		},
	}
}

// GetConfiguration - Get configuration
func (c *ConfigurationAPIController) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.GetConfiguration(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}

// PutConfiguration - Updates the configuration
func (c *ConfigurationAPIController) PutConfiguration(w http.ResponseWriter, r *http.Request) {
	var configurationParam Configuration
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&configurationParam); err != nil && !errors.Is(err, io.EOF) {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertConfigurationRequired(configurationParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertConfigurationConstraints(configurationParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.PutConfiguration(r.Context(), configurationParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	_ = EncodeJSONResponse(result.Body, &result.Code, w)
}
