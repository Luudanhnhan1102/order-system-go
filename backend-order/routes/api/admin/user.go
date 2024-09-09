package admin

import (
	"context"
	"net/http"

	"backend-order/database"
	"backend-order/middleware"
	"backend-order/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetupAdminUserRoutes(r *gin.Engine) {
	// Add admin routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
	{
		adminGroup.GET("/users", listUsersHandler)
		adminGroup.GET("/users/:id", getUserDetailsHandler)
		adminGroup.POST("/users", createUserHandler)
		adminGroup.PUT("/users/:id", updateUserHandler)
	}
}

// listUsersHandler handles retrieving the list of users for admin
// @Summary List all users
// @Description Retrieve a list of all users (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users [get]
func listUsersHandler(c *gin.Context) {
	var users []models.User
	ctx := context.Background()

	db := database.GetDB()
	err := db.Collection("users").Find(ctx, bson.M{}).All(&users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// getUserDetailsHandler handles retrieving detailed information of a specific user for admin
// @Summary Get user details
// @Description Retrieve detailed information of a specific user (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{id} [get]
func getUserDetailsHandler(c *gin.Context) {
	// Extract user ID from URL params
	userID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Fetch user from database
	var user models.User
	ctx := context.Background()
	db := database.GetDB()
	err = db.Collection("users").Find(ctx, bson.M{"_id": userID}).One(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return user details
	c.JSON(http.StatusOK, user)
}

// createUserHandler handles creating a new user by admin
// @Summary Create a new user
// @Description Create a new user (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param user body models.User true "User object"
// @Security ApiKeyAuth
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users [post]
func createUserHandler(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use SetPassword method
	if err := newUser.SetPassword(newUser.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}

	ctx := context.Background()
	db := database.GetDB()
	result, err := db.Collection("users").InsertOne(ctx, &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	newUser.ID = result.InsertedID.(primitive.ObjectID)
	newUser.Password = "" // Don't send the password back
	c.JSON(http.StatusCreated, newUser)
}

// updateUserHandler handles updating an existing user by admin
// @Summary Update an existing user
// @Description Update an existing user (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.User true "Updated user object"
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users/{id} [put]
func updateUserHandler(c *gin.Context) {
	userID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var updateUser models.User
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	db := database.GetDB()

	// If password is provided, use SetPassword method
	if updateUser.Password != "" {
		if err := updateUser.SetPassword(updateUser.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
			return
		}
	}

	updateData := bson.M{
		"email":    updateUser.Email,
		"password": updateUser.Password, // Only if password was updated
	}

	// Remove empty fields
	for k, v := range updateData {
		if v == "" {
			delete(updateData, k)
		}
	}

	err = db.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": updateData})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	updateUser.Password = "" // Don't send the password back
	c.JSON(http.StatusOK, updateUser)
}
