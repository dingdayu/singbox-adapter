// Package middleware defines gin middlewares.
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// RequestLogNamed is the logger namespace for HTTP requests.
	RequestLogNamed    = "http_request"
	maxBodyLogSize     = 1024 // max request body size to log (1KB)
	maxResponseLogSize = 1024
)

// Some response content types should not be logged as body
var downloadableContentTypes = []string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // Excel file
	"application/pdf",          // PDF file
	"application/octet-stream", // generic binary
	"image/jpeg",               // JPEG image
	"image/png",                // PNG image
	"image/gif",                // GIF image
	"image/bmp",                // BMP image
	"image/webp",               // WEBP image
	"text/csv",                 // CSV file
	"text/html",                // HTML file
	"text/javascript",          // JavaScript file
	"application/javascript",   // JavaScript file
	"text/css",                 // CSS file
	"font/ttf",                 // TrueType font
	"image/svg+xml",            // SVG file
	"application/zip",          // ZIP file
	"application/x-rar-compressed",
	"application/x-7z-compressed",
	"application/x-bzip2",
	"application/x-bzip",
	"application/x-gzip",
	// add more file types here
}

// WriterLog logs request/response with masking of sensitive fields.
func WriterLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// capture request body for logging
		bodyBuf := new(bytes.Buffer)
		_, _ = io.Copy(bodyBuf, c.Request.Body)
		body := bodyBuf.Bytes()
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		contentType := c.Writer.Header().Get("Content-Type")

		fs := []any{
			slog.Int("status", c.Writer.Status()),
			slog.String("ip", c.ClientIP()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int64("latency", latency.Milliseconds()),
			slog.String("user_agent", c.Request.UserAgent()),
		}
		// log query string when present
		if len(c.Request.URL.RawQuery) > 0 {
			fs = append(fs, slog.String("query", c.Request.URL.RawQuery))
		}
		// avoid logging non-textual or large responses
		if !isFileResponse(contentType) && blw.body.Len() <= maxResponseLogSize {
			fs = append(fs, slog.String("response", string(mask(blw.body.Bytes()))))
		}

		// Append error field if this is an erroneous request.
		if len(c.Errors) > 0 {
			fs = append(fs, slog.String("errors", c.Errors.String()))
		}
		// log request body for non-GET when size <= 1KB
		if c.Request.Method != http.MethodGet && len(body) <= maxBodyLogSize {
			fs = append(fs, slog.Any("body", mask(body)))
		}
		// write X-Request-ID to log
		xRequestID := c.Request.Header.Get("X-Request-Id")
		if len(xRequestID) > 0 {
			fs = append(fs, slog.String("request_id", xRequestID))
		}

		logger.InfoContext(c.Request.Context(), c.Request.RequestURI, fs...)
	}
}

var maskDictionary = map[string]bool{"password": true}

func mask(body []byte) []byte {
	// unmarshal JSON to map
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}

	// mask sensitive fields
	filterSensitiveData(data, maskDictionary)

	// marshal back to JSON
	filteredJSON, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return filteredJSON
}

// filterSensitiveData recursively masks keys found in maskDictionary.
func filterSensitiveData(data map[string]interface{}, maskDictionary map[string]bool) {
	for key, value := range data {
		// mask when key is sensitive
		if maskDictionary[key] {
			data[key] = "***"
		} else {
			// recursively mask nested structures
			switch v := value.(type) {
			case map[string]interface{}:
				filterSensitiveData(v, maskDictionary)
			case []interface{}:
				for _, item := range v {
					if itemMap, ok := item.(map[string]interface{}); ok {
						filterSensitiveData(itemMap, maskDictionary)
					}
				}
			}
		}
	}
}

// isFileResponse reports whether contentType matches downloadable content types.
func isFileResponse(contentType string) bool {
	for _, fileType := range downloadableContentTypes {
		if strings.Contains(contentType, fileType) {
			return true
		}
	}
	return false
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
