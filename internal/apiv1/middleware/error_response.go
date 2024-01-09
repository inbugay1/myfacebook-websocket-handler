package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/inbugay1/httprouter"
	"myfacebook-websocket-handler/internal/apiv1"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type errorResponse struct {
	next httprouter.Handler
}

func (m *errorResponse) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	err := m.next.Handle(responseWriter, request)

	if err == nil {
		return nil
	}

	apiError, ok := err.(*apiv1.Error) //nolint:errorlint
	if !ok {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return nil
	}

	err = m.sendJSONError(responseWriter, apiError)
	if err != nil {
		return fmt.Errorf("middleware, errorResponse.Handle, m.sendJSONError, err: %w", err)
	}

	return nil
}

func (m *errorResponse) sendJSONError(responseWriter http.ResponseWriter, apiErr *apiv1.Error) error {
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.WriteHeader(apiErr.StatusCode())

	errorResponse := ErrorResponse{
		Message: apiErr.Message(),
		Code:    apiErr.Code(),
	}

	err := json.NewEncoder(responseWriter).Encode(errorResponse)
	if err != nil {
		return fmt.Errorf("json.NewEncoder, err: %w", err)
	}

	return nil
}

func NewErrorResponse() httprouter.MiddlewareFunc {
	return func(next httprouter.Handler) httprouter.Handler {
		return &errorResponse{next: next}
	}
}
