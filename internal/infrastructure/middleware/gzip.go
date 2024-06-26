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
	gzipString            = "gzip"
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
		return n, fmt.Errorf("ошибка при записи ответа для compressWriter: %w", err)
	}
	return n, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < statusMultipleChoices {
		c.w.Header().Set("Content-Encoding", gzipString)
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("ошибка при закрытии ответа для compressWriter: %w", err)
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
		return nil, fmt.Errorf("ошибка при чтении запроса при инициализации compressReader: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	n, err = c.zr.Read(p)
	if err != nil {
		return n, fmt.Errorf("ошибка при чтении запроса для compressReader: %w", err)
	}
	return n, nil
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("ошибка при закрытии чтения для compressReader: %w", err)
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

		supportsGzip := strings.Contains(acceptEncoding, gzipString)
		supportApplicationJSON := strings.Contains(contentType, applicationJSON)
		supportTextHTML := strings.Contains(accept, textHTML)

		if supportsGzip && (supportApplicationJSON || supportTextHTML) {
			cw := newCompressWriter(w)
			ow = cw

			w.Header().Set("Content-Encoding", gzipString)

			defer func() {
				err := cw.Close()
				if err != nil {
					gm.logger.Info("ошибка при закрытии compressWriter", zap.Error(err))
				}
			}()
		}

		contentEncoding := strings.Join(r.Header.Values("Content-Encoding"), ", ")
		sendsGzip := strings.Contains(contentEncoding, gzipString)
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer func() {
				err := cr.Close()
				if err != nil {
					gm.logger.Info("ошибка при закрытии compressReader", zap.Error(err))
				}
			}()
		}

		next.ServeHTTP(ow, r)
	})
}
