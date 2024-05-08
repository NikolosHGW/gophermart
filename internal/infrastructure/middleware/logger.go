package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type LoggerInterface interface {
	Sugar() *zap.SugaredLogger
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return size, fmt.Errorf("ошибка при записи ответа: %w", err)
	}
	r.responseData.size += size

	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

type LoggerMiddleware struct {
	logger LoggerInterface
}

func NewLoggerMiddleware(logger LoggerInterface) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

func (lm LoggerMiddleware) WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		lm.logger.Sugar().Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	}

	return http.HandlerFunc(logFn)
}
