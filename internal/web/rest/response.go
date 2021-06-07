package rest

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/http"
)

type Response struct {
	StatusCode   int         `json:"status_code"`
	ErrorMessage string      `json:"error_message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

type responseWriter struct {
	writer     http.ResponseWriter
	router     *router
	statusCode int
}

func wrapResponseWriter(writer http.ResponseWriter, router *router) *responseWriter {
	return &responseWriter{
		writer: writer,
		router: router,
	}
}

type HandlerFunc func(*responseWriter, *http.Request)

func (r *router) wrapStandardHttpMethod(handlerFunc HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handlerFunc(wrapResponseWriter(writer, r), request)
	}
}

func (writer responseWriter) Header() http.Header {
	return writer.writer.Header()
}

func (writer *responseWriter) Write(bytes []byte) (int, error) {
	if writer.statusCode == 0 {
		writer.statusCode = http.StatusOK
	}
	return writer.writer.Write(bytes)
}

func (writer responseWriter) WriteHeader(statusCode int) {
	writer.writer.WriteHeader(statusCode)
	writer.statusCode = statusCode
}

func (writer responseWriter) WriteResponse(statusCode int, errorMessage string, data interface{}, request *http.Request) {
	writer.WriteHeader(statusCode)
	resp := &Response{
		StatusCode:   statusCode,
		ErrorMessage: errorMessage,
	}
	if data != nil {
		resp.Data = data
	}
	if err := json.NewEncoder(writer.writer).Encode(resp); err != nil {
		writer.router.log(zerolog.ErrorLevel, request).Err(err).Msg("could not write http response")
		writer.WriteAutomaticErrorResponse(http.StatusInternalServerError, nil, request)
	}
}

func (writer responseWriter) WriteSuccessfulResponse(data interface{}, r *http.Request) {
	writer.WriteResponse(http.StatusOK, "", data, r)
}

func (writer responseWriter) WriteNotFoundResponse(message string, data interface{}, r *http.Request) {
	writer.WriteResponse(http.StatusOK, message, data, r)
}

func (writer responseWriter) WriteAutomaticErrorResponse(statusCode int, data interface{}, r *http.Request) {
	writer.WriteResponse(statusCode, http.StatusText(statusCode), data, r)
}
