package handler

import (
	"context"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yamasaki/static-server/internal/storage"
)

type StaticHandler struct {
	storage *storage.MinioStorage
	logger  *logrus.Logger
}

func NewStaticHandler(storage *storage.MinioStorage, logger *logrus.Logger) *StaticHandler {
	return &StaticHandler{
		storage: storage,
		logger:  logger,
	}
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	
	bucket := extractBucketFromHost(host)
	if bucket == "" {
		h.logger.WithField("host", host).Error("Could not extract bucket from host")
		http.Error(w, "Invalid host", http.StatusBadRequest)
		return
	}

	objectPath := strings.TrimPrefix(r.URL.Path, "/")
	if objectPath == "" || strings.HasSuffix(objectPath, "/") {
		objectPath = path.Join(objectPath, "index.html")
	}

	ctx := context.Background()
	
	if !h.storage.ObjectExists(ctx, bucket, objectPath) {
		if !strings.HasSuffix(objectPath, ".html") {
			htmlPath := objectPath + ".html"
			if h.storage.ObjectExists(ctx, bucket, htmlPath) {
				objectPath = htmlPath
			} else {
				h.logger.WithFields(logrus.Fields{
					"bucket": bucket,
					"object": objectPath,
				}).Warn("Object not found")
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}
		} else {
			h.logger.WithFields(logrus.Fields{
				"bucket": bucket,
				"object": objectPath,
			}).Warn("Object not found")
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}

	object, err := h.storage.GetObject(ctx, bucket, objectPath)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get object from storage")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer object.Close()

	contentType := getContentType(objectPath)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")

	if _, err := io.Copy(w, object); err != nil {
		h.logger.WithError(err).Error("Failed to write response")
	}
}

func extractBucketFromHost(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}
	return ""
}

func getContentType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".webp":
		return "image/webp"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}