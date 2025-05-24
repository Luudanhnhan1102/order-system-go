package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"backend-order/database"
	"backend-order/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

// IPRateLimiter holds the rate limiter for each IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu   *sync.RWMutex
	r    rate.Limit // requests per second
	b    int       // burst size
}

// NewIPRateLimiter creates a new IP rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:   &sync.RWMutex{},
		r:    r,
		b:    b,
	}
	return i
}

// AddIP creates a new rate limiter and adds it to the ips map
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise, calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		return i.AddIP(ip)
	}

	return limiter
}

// RateLimiter is the global rate limiter instance
var RateLimiter = NewIPRateLimiter(1, 5) // Allow 1 request per second with a burst of 5

var (
	jwtSecretKey  []byte
	secretKeyOnce sync.Once
)

// GetSecretKey returns the JWT secret key, initializing it if necessary
func GetSecretKey() ([]byte, error) {
	var initErr error
	secretKeyOnce.Do(func() {
		key := os.Getenv("JWT_SECRET_KEY")
		if key == "" {
			initErr = fmt.Errorf("JWT_SECRET_KEY environment variable not set")
			return
		}
		// Ensure the key is at least 32 bytes for HS256
		if len(key) < 32 {
			key = key + strings.Repeat("0", 32-len(key))
			key = key[:32] // Ensure it's exactly 32 bytes
		}
		jwtSecretKey = []byte(key)
		log.Printf("JWT Secret Key loaded, length: %d", len(jwtSecretKey))
	})
	return jwtSecretKey, initErr
}

// RateLimitMiddleware applies rate limiting based on IP
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := RateLimiter.GetLimiter(c.ClientIP())

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the secret key
		key, err := GetSecretKey()
		if err != nil {
			log.Printf("Error getting JWT secret key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		log.Printf("Processing auth request for path: %s", c.Request.URL.Path)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		// Extract the token from the Bearer string
		tokenString := ""
		const bearerPrefix = "Bearer "
		if strings.HasPrefix(authHeader, bearerPrefix) {
			tokenString = strings.TrimPrefix(authHeader, bearerPrefix)
		} else {
			log.Println("Invalid token format. Use 'Bearer <token>'")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format. Use 'Bearer <token>'"})
			c.Abort()
			return
		}

		log.Printf("Extracted token, length: %d", len(tokenString))
		log.Printf("JWT Secret key length: %d", len(key))

		log.Printf("JWT Secret key length: %d", len(key))

		// Parse the token with the secret key
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			log.Printf("Using signing method: %v", token.Method.Alg())
			log.Printf("Token header: %+v", token.Header)
			// Convert key to string and back to byte slice to ensure consistency
			keyStr := string(key)
			return []byte(keyStr), nil
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil {
			log.Printf("Token validation error: %v", err)
			log.Printf("Token string: %s", tokenString)
		}

		if err != nil || token == nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(claims["id"].(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		db := database.GetDB()
		var user models.User
		err = db.Collection("users").Find(context.Background(), bson.M{"_id": userID}).One(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userModel, ok := user.(models.User)
		if !ok || !userModel.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
