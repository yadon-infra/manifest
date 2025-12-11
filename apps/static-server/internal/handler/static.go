package handler

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
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

func (h *StaticHandler) ServeFiles(c *gin.Context) {
	host := c.Request.Host
	bucket := extractBucketFromHost(host)
	if bucket == "" {
		h.logger.WithField("host", host).Error("Could not extract bucket from host")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host"})
		return
	}
	objectPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	if strings.HasSuffix(objectPath, "/") || objectPath == "" {
		objectPath = path.Join(objectPath, "index.html")
	}
	// Add bucket name as prefix since objects are stored with bucket prefix
	objectPath = path.Join(bucket, objectPath)

	h.logger.WithFields(logrus.Fields{
		"bucket":     bucket,
		"objectPath": objectPath,
		"host":       host,
		"url":        c.Request.URL.Path,
	}).Debug("Attempting to serve file")

	if !h.storage.ObjectExists(c, bucket, objectPath) {
		if !strings.HasSuffix(objectPath, ".html") {
			htmlPath := objectPath + ".html"
			if h.storage.ObjectExists(c, bucket, htmlPath) {
				objectPath = htmlPath
			} else {
				h.logger.WithFields(logrus.Fields{
					"bucket": bucket,
					"object": objectPath,
				}).Warn("Object not found")
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}
		} else {
			h.logger.WithFields(logrus.Fields{
				"bucket": bucket,
				"object": objectPath,
			}).Warn("Object not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
			return
		}
	}

	object, err := h.storage.GetObject(c, bucket, objectPath)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"bucket": bucket,
			"object": objectPath,
		}).Error("Failed to get object from storage")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer object.Close()
	contentType := getContentType(objectPath)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600")
	if _, err := io.Copy(c.Writer, object); err != nil {
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
