package middleware

type Middlewares struct {
	Logger *LoggerMiddleware
	Gzip   *GzipMiddleware
}
