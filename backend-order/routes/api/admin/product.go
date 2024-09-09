package admin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"backend-order/database"
	"backend-order/middleware"
	"backend-order/models"
)

func SetupAdminProductRoutes(r *gin.Engine) {
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
	{
		adminGroup.POST("/products", CreateProduct)
		adminGroup.DELETE("/products/:id", DeleteProduct)
	}
}

// ProductInput defines the structure for product creation input
type ProductInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	// Add other fields as needed
}

// CreateProduct godoc
// @Summary Create a new product zzz
// @Description Create a new product (admin only)
// @Tags admin,products
// @Accept json
// @Produce json
// @Param product body ProductInput true "Product information"
// @Success 201 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/products [post]
func CreateProduct(c *gin.Context) {
	var newProduct models.Product

	// Bind JSON body to the newProduct struct
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
		return
	}

	// Generate a new ObjectID for the product
	newProduct.ID = primitive.NewObjectID()

	// Insert the new product into the database
	ctx := context.Background()
	db := database.GetDB()
	_, err := db.Collection("products").InsertOne(ctx, newProduct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Return the created product
	c.JSON(http.StatusCreated, newProduct)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product by ID (admin only)
// @Tags admin,products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	result, err := db.Collection("products").RemoveAll(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
