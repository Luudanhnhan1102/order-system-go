package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"backend-order/database"
	"backend-order/models"
	"backend-order/vendors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var (
	secretKey = []byte("your_secret_key") // TODO: Use environment variable in production
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	ResetToken  string `json:"resetToken" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func SetupAuthRoutes(r *gin.Engine) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", loginHandler)
		authGroup.POST("/register", registerUserHandler)
		authGroup.POST("/reset-password", resetPasswordHandler)   // New endpoint
		authGroup.POST("/forgot-password", forgotPasswordHandler) // New endpoint
	}
}

// loginHandler handles user authentication and JWT token generation
// @Summary User login
// @Description Authenticate a user and return a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func loginHandler(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := authenticateUser(loginReq.Email, loginReq.Password)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	token, err := generateJWTToken(*user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// registerUserHandler handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param registerRequest body RegisterUserRequest true "User registration details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func registerUserHandler(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	db := database.GetDB()



	var existingUser models.User
	err := db.Collection("users").Find(c, primitive.M{"email": req.Email}).One(&existingUser)

		// log the user email
		fmt.Println("User email:", req.Email)
		fmt.Println("User password:", req.Password)
		fmt.Println("User existingUser:", existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
		return
	}

	user := models.User{
		ID:      primitive.NewObjectID(),
		Email:   req.Email,
		IsAdmin: false, // Set new users as non-admin by default
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}

	_, err = db.Collection("users").InsertOne(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Send welcome email
	emailData := vendors.EmailData{
		To:      []vendors.EmailAddress{{Email: user.Email, Name: user.Email}},
		Subject: "Welcome to Our Service",
		Text:    fmt.Sprintf("Dear %s,\n\nWelcome to our service! Your account has been successfully created.\n\nBest regards,\nThe Team", user.Email),
	}

	err = vendors.SendEmail(emailData)
	if err != nil {
		// Log the error, but don't return it to the user
		fmt.Printf("Failed to send welcome email: %v\n", err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
}

// resetPasswordHandler handles password reset requests
// @Summary Reset user password
// @Description Reset a user's password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param resetRequest body ResetPasswordRequest true "Password reset details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/reset-password [post]
func resetPasswordHandler(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User
	err := db.Collection("users").Find(c, bson.M{
		"email":         req.Email,
		"resetToken":    req.ResetToken,
		"resetTokenExp": bson.M{"$gt": time.Now()},
	}).One(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set new password"})
		return
	}

	err = db.Collection("users").UpdateOne(
		c,
		bson.M{"_id": user.ID},
		bson.M{
			"$set":   bson.M{"password": user.Password},
			"$unset": bson.M{"resetToken": "", "resetTokenExp": ""},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password in database"})
		return
	}

	// Send password reset confirmation email
	emailData := vendors.EmailData{
		To:      []vendors.EmailAddress{{Email: user.Email, Name: user.Email}},
		Subject: "Password Reset Successful",
		Text:    fmt.Sprintf("Dear %s,\n\nYour password has been successfully reset. If you did not initiate this change, please contact our support team immediately.\n\nBest regards,\nThe Team", user.Email),
	}

	err = vendors.SendEmail(emailData)
	if err != nil {
		// Log the error, but don't return it to the user
		fmt.Printf("Failed to send password reset confirmation email: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// forgotPasswordHandler initiates the forgot password process
// @Summary Initiate forgot password process
// @Description Send a password reset token to the user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param forgotRequest body ForgotPasswordRequest true "User's email"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/forgot-password [post]
func forgotPasswordHandler(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User
	err := db.Collection("users").Find(c, bson.M{"email": req.Email}).One(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate a reset token
	resetToken, err := generateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	// Store the reset token in the database
	err = db.Collection("users").UpdateOne(
		c,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"resetToken":    resetToken,
				"resetTokenExp": time.Now().Add(15 * time.Minute), // Token expires in 15 minutes
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store reset token"})
		return
	}

	// Send email to the user with the reset token
	emailData := vendors.EmailData{
		To:      []vendors.EmailAddress{{Email: user.Email, Name: user.Email}},
		Subject: "Password Reset Request",
		Text:    fmt.Sprintf("Your password reset token is: %s\nThis token will expire in 15 minutes.", resetToken),
	}

	err = vendors.SendEmail(emailData)
	if err != nil {
		fmt.Println("Failed to send reset token email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset token email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset initiated. Check your email for further instructions.",
	})
}

func generateResetToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func authenticateUser(email, password string) (*models.User, error) {
	db := database.GetDB()
	var user models.User

	filter := bson.M{"email": email}
	err := db.Collection("users").Find(context.Background(), filter).One(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	return &user, nil
}

func generateJWTToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":      user.ID.Hex(),
		"email":   user.Email,
		"isAdmin": user.IsAdmin,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(secretKey)
}
