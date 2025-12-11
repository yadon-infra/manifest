package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yamasaki/static-server/internal/handler"
	"github.com/yamasaki/static-server/internal/storage"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	endpoint := getEnv("MINIO_ENDPOINT", "minio.minio.svc.cluster.local:9000")
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
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	
	r.Use(loggingMiddleware(logger))
	r.Use(gin.Recovery())
	
	// Health check route
	r.GET("/", staticHandler.ServeFiles)
	r.NoRoute(staticHandler.ServeFiles)

	logger.WithField("port", port).Info("Starting server")
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		logger.WithError(err).Fatal("Server failed")
	}
}

func loggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"host":       c.Request.Host,
			"status":     c.Writer.Status(),
			"duration":   time.Since(start).Milliseconds(),
			"user_agent": c.Request.UserAgent(),
		}).Info("Request handled")
	}
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