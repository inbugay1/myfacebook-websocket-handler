package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/inbugay1/httprouter"
	"myfacebook-websocket-handler/internal/apiv1"
)

type errorLog struct {
	next httprouter.Handler
}

const ErrorLogFormat = "%s: %s"

func (m *errorLog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	err := m.next.Handle(responseWriter, request)
	if err != nil {
		m.logError(err)
	}

	return err //nolint:wrapcheck
}

func (m *errorLog) logError(err error) {
	errorMessage := fmt.Sprintf(ErrorLogFormat, reflect.TypeOf(err), err)

	apiError, ok := err.(*apiv1.Error) //nolint:errorlint
	if !ok {
		slog.Error(errorMessage)

		return
	}

	switch apiError.LogLevel() {
	case apiv1.ErrorLogLevelInfo:
		slog.Info(errorMessage)
	case apiv1.ErrorLogLevelWarning:
		slog.Warn(errorMessage)
	default:
		slog.Error(errorMessage)
	}
}

func NewErrorLog() httprouter.MiddlewareFunc {
	return func(next httprouter.Handler) httprouter.Handler {
		return &errorLog{next: next}
	}
}
