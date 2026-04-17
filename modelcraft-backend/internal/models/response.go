package models

import "net/http"

// TokenRequest represents the request body for token exchange
type TokenRequest struct {
	Code string `json:"code" binding:"required"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description,omitempty"`
	Code        string `json:"code,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewErrorResponse creates a new ErrorResponse
func NewErrorResponse(error, description, code string) *ErrorResponse {
	return &ErrorResponse{
		Error:       error,
		Description: description,
		Code:        code,
	}
}

// NewSuccessResponse creates a new SuccessResponse
func NewSuccessResponse(data interface{}, message string) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// ErrorResponseWithStatus creates an ErrorResponse with HTTP status code
func ErrorResponseWithStatus(error, description, code string, status int) (interface{}, int) {
	return NewErrorResponse(error, description, code), status
}

// SuccessResponseWithStatus creates a SuccessResponse with HTTP status code
func SuccessResponseWithStatus(data interface{}, message string, status int) (interface{}, int) {
	return NewSuccessResponse(data, message), status
}

// HTTP status code aliases for convenience
const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusInternalServerError = http.StatusInternalServerError
)
