package api

import (
	"context"
	"net/http"

	"backend-order/database"
	"backend-order/models"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SetupProductRoutes sets up the product-related routes
func SetupProductRoutes(r *gin.Engine) {
	r.GET("/products", getProductsHandler)
	r.GET("/products/:id", GetProduct)
}

// @Summary Get products
// @Description Get the list of all products
// @Tags Products
// @Produce json
// @Success 200 {array} models.Product
// @Router /products [get]
func getProductsHandler(c *gin.Context) {
	var products []models.Product
	ctx := context.Background()

	db := database.GetDB()
	err := db.Collection("products").Find(ctx, bson.M{}).All(&products)
	if err != nil {
		// If the error is due to no documents found, return an empty array
		if err == qmgo.ErrNoSuchDocuments {
			c.JSON(http.StatusOK, []models.Product{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching products"})
		return
	}

	// If products is nil, return an empty array instead
	if products == nil {
		products = []models.Product{}
	}

	c.JSON(http.StatusOK, products)
}

// @Summary Get product by ID
// @Description Get detailed information of a specific product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /products/{id} [get]
func GetProduct(c *gin.Context) {
	productID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	db := database.GetDB()
	var product models.Product

	err = db.Collection("products").Find(context.Background(), bson.M{"_id": objID}).One(&product)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}
