package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/inbugay1/httprouter"
)

type requestResponseLog struct {
	next httprouter.Handler
}

const RequestResponseLogFormat = "%d [%s] %s: Request Time: %s Response Time: %s Request Headers: %s Request Body: %s Response Headers: %+v Response Body: %s"

func (m *requestResponseLog) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	requestDateTime := time.Now()

	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("middleware, requestResponseLog.Handle, io.ReadAll, err: %w", err)
	}

	request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	responseWriterRecorder := httptest.NewRecorder()

	// passing already set headers to responseWriterRecorder
	for k, v := range responseWriter.Header() {
		responseWriterRecorder.Header()[k] = v
	}

	err = m.next.Handle(responseWriterRecorder, request)
	if err != nil {
		return fmt.Errorf("middleware, requestResponseLog.Handle, m.next.Handle, err: %w", err)
	}

	responseDateTime := time.Now()

	resultRecorder := responseWriterRecorder.Result()

	responseBody, err := io.ReadAll(resultRecorder.Body)
	if err != nil {
		return fmt.Errorf("middleware, requestResponseLog.Handle, ioutil.ReadAll, err: %w", err)
	}

	slog.Debug(fmt.Sprintf(RequestResponseLogFormat,
		resultRecorder.StatusCode,
		request.Method,
		request.URL.RequestURI(),
		requestDateTime,
		responseDateTime,
		request.Header,
		requestBody,
		resultRecorder.Header,
		responseBody,
	))

	// Send data from recorder to http response, do not change order
	for k, v := range resultRecorder.Header {
		responseWriter.Header()[k] = v
	}

	responseWriter.WriteHeader(resultRecorder.StatusCode)

	_, err = responseWriterRecorder.Body.WriteTo(responseWriter)
	if err != nil {
		return fmt.Errorf("middleware, requestResponseLog.Handle, responseWriterRecorder.Body.WriteTo, err: %w", err)
	}

	err = resultRecorder.Body.Close()
	if err != nil {
		return fmt.Errorf("middleware, requestResponseLog.Handle, resultRecorder.Body.Close, err: %w", err)
	}

	return nil
}

func NewRequestResponseLog() httprouter.MiddlewareFunc {
	return func(next httprouter.Handler) httprouter.Handler {
		return &requestResponseLog{next: next}
	}
}
