package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	SignatureHeader = "X-Signature"
	TimestampHeader = "X-Timestamp"
)

var secretKey = []byte(os.Getenv("API_SECRET_KEY"))

func SignRequest(method, path string, body []byte, timestamp time.Time) string {
	message := fmt.Sprintf("%s%s%s%d", method, path, body, timestamp.Unix())
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get(SignatureHeader)
	timestamp := r.Header.Get(TimestampHeader)
	fmt.Printf("signature: %s, timestamp: %s\n", signature, timestamp)
	if signature == "" || timestamp == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return false
	}

	// Check if the request is not older than 5 minutes
	if time.Since(t) > 5*time.Minute {
		return false
	}

	expectedSignature := SignRequest(r.Method, r.URL.Path, body, t)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
