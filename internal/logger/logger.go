package logger

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Initialize() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
		FieldsOrder: []string{
			"type", "method", "url", "status", "request_id", "duration", "bytes_in", "bytes_out", "error", "errors",
		},
		TimeFormat: time.RFC3339,
	})
}

func Configure(config *config.LogConfig) error {
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)
	return nil
}

var LoggerContextKey = "logger"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := middleware.GetReqID(ctx)
		logger := log.Logger.With().Str("request_id", requestID).Logger()
		ctx = context.WithValue(ctx, LoggerContextKey, &logger)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		defer func() {
			logger.Info().Str("type", "request").Str("method", r.Method).Str("url", r.URL.String()).Int("status", ww.Status()).Dur("duration", time.Since(start)).Int64("bytes_in", r.ContentLength).Int("bytes_out", ww.BytesWritten()).Send()
		}()
		next.ServeHTTP(ww, r.WithContext(ctx))
	})
}

func Get(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &log.Logger
	}
	if logger, ok := ctx.Value(LoggerContextKey).(*zerolog.Logger); ok {
		return logger
	}
	return &log.Logger
}
