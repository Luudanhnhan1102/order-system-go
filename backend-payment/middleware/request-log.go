package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type logEntry struct {
	Timestamp      string            `json:"timestamp"`
	Method         string            `json:"method"`
	Path           string            `json:"path"`
	RequestHeader  map[string]string `json:"request_header"`
	QueryParams    map[string]string `json:"query_params"`
	RequestBody    string            `json:"request_body"`
	Status         int               `json:"status"`
	ResponseHeader map[string]string `json:"response_header"`
	ResponseBody   string            `json:"response_body"`
	Duration       string            `json:"duration"`
	ClientIP       string            `json:"client_ip"`
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a new response writer
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Prepare log entry
		entry := logEntry{
			Timestamp:      time.Now().Format(time.RFC3339),
			Method:         c.Request.Method,
			Path:           c.Request.URL.Path,
			RequestHeader:  make(map[string]string),
			QueryParams:    make(map[string]string),
			RequestBody:    string(requestBody),
			Status:         c.Writer.Status(),
			ResponseHeader: make(map[string]string),
			ResponseBody:   blw.body.String(),
			Duration:       time.Since(start).String(),
			ClientIP:       c.ClientIP(),
		}

		// Copy headers (converting []string to string)
		for k, v := range c.Request.Header {
			entry.RequestHeader[k] = v[0]
		}
		for k, v := range c.Writer.Header() {
			entry.ResponseHeader[k] = v[0]
		}

		// Copy query parameters
		for k, v := range c.Request.URL.Query() {
			entry.QueryParams[k] = v[0]
		}

		// Convert to JSON and print
		jsonEntry, err := json.Marshal(entry)
		if err == nil {
			println(string(jsonEntry))
		}
	}
}
