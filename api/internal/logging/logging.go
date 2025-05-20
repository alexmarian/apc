package logging

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var Logger *zap.Logger

// Initialize sets up the zap logger
func Initialize(logLevel string, logFilePath string, isDevelopment bool) error {
	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	if isDevelopment {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Configure output
	var output zapcore.WriteSyncer
	if logFilePath != "" {
		// Ensure directory exists
		dir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		output = zapcore.AddSync(file)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create the logger
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		output,
		level,
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

// Helper functions that add request context
func WithRequestContext(requestID, userID, path string) *zap.Logger {
	return Logger.With(
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("path", path),
	)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()
		userID := "anonymous"
		if id, ok := r.Context().Value("userID").(string); ok {
			userID = id
		}

		requestLogger := WithRequestContext(requestID, userID, r.URL.Path)
		clientIP := r.RemoteAddr
		forwardedFor := r.Header.Get("X-Forwarded-For")
		if forwardedFor != "" {
			clientIP = forwardedFor
		}
		requestLogger.Info("Request started",
			zap.String("method", r.Method),
			zap.String("remote_addr", clientIP),
			zap.String("user_agent", r.UserAgent()),
		)
		ww := NewResponseWriter(w)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		requestLogger.Info("Request completed",
			zap.Int("status", ww.Status()),
			zap.Duration("duration", duration),
		)
	})
}

// ResponseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Status() int {
	return rw.status
}
