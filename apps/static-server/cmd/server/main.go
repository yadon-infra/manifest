package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/yamasaki/static-server/internal/handler"
	"github.com/yamasaki/static-server/internal/storage"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	endpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	accessKey := getEnv("MINIO_ACCESS_KEY", "")
	secretKey := getEnv("MINIO_SECRET_KEY", "")
	useSSL := getEnvBool("MINIO_USE_SSL", false)
	port := getEnv("PORT", "8080")

	if accessKey == "" || secretKey == "" {
		logger.Fatal("MINIO_ACCESS_KEY and MINIO_SECRET_KEY must be set")
	}

	minioStorage, err := storage.NewMinioStorage(endpoint, accessKey, secretKey, useSSL, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize MinIO storage")
	}

	staticHandler := handler.NewStaticHandler(minioStorage, logger)

	r := mux.NewRouter()
	r.PathPrefix("/").Handler(staticHandler)

	srv := &http.Server{
		Handler:      loggingMiddleware(logger)(r),
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.WithField("port", port).Info("Starting server")
	if err := srv.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("Server failed")
	}
}

func loggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			
			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"host":       r.Host,
				"status":     wrapped.statusCode,
				"duration":   time.Since(start).Milliseconds(),
				"user_agent": r.UserAgent(),
			}).Info("Request handled")
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return b
	}
	return defaultValue
}