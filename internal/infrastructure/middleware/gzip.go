package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	applicationJSON       = "application/json"
	textHTML              = "html/text"
	statusMultipleChoices = 300
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return n, fmt.Errorf("ошибка при записи ответа: %w", err)
	}
	return n, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < statusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("ошибка при закрытии ответа: %w", err)
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении запроса: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	n, err = c.zr.Read(p)
	if err != nil {
		return n, fmt.Errorf("ошибка при чтении запроса: %w", err)
	}
	return n, nil
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("ошибка при закрытии чтения: %w", err)
	}
	return nil
}

type ZapLogger interface {
	Info(msg string, fields ...zapcore.Field)
}

type GzipMiddleware struct {
	logger ZapLogger
}

func NewGzipMiddleware(logger ZapLogger) *GzipMiddleware {
	return &GzipMiddleware{
		logger: logger,
	}
}

func (gm GzipMiddleware) WithGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := strings.Join(r.Header.Values("Accept-Encoding"), ", ")
		contentType := strings.Join(r.Header.Values("Content-Type"), ", ")
		accept := strings.Join(r.Header.Values("Accept"), ", ")

		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		supportApplicationJSON := strings.Contains(contentType, applicationJSON)
		supportTextHTML := strings.Contains(accept, textHTML)

		if supportsGzip && (supportApplicationJSON || supportTextHTML) {
			cw := newCompressWriter(w)
			ow = cw

			defer func() {
				if err := cw.Close(); err != nil {
					gm.logger.Info("ошибка при закрытии compressWriter: %v\n", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}()
		}

		contentEncoding := strings.Join(r.Header.Values("Content-Encoding"), ", ")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					gm.logger.Info("ошибка при закрытии compressReader: %v\n", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}()
		}

		next.ServeHTTP(ow, r)
	})
}
