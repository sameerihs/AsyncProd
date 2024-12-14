package handlers

import (
	"AsyncProd/models"
	"AsyncProd/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// handles new product creation
func CreateProductHandler(c *gin.Context) {
    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    productID, err := models.SaveProduct(&product)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
        return
    }

    // publish message for image processing
    err = services.PublishImageProcessingMessage(productID, product.UserID, product.ProductImages)
    if err != nil {
        log.Printf("ERROR: Failed to publish image processing message: %v", err)
    } else {
        log.Printf("SUCCESS: Published image processing message for product ID: %d", productID)
    }

    c.JSON(http.StatusCreated, gin.H{"product_id": productID})
}

// retrieves a product by ID

func GetProductByIDHandler(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    product, err := models.GetProductByID(id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, product)
}

// updates a product
func UpdateProductHandler(c *gin.Context) {
    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    if err := models.UpdateProduct(&product); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Publish message for image processing
    err := services.PublishImageProcessingMessage(product.ID, product.UserID, product.ProductImages)
    if err != nil {
        log.Printf("ERROR: Failed to publish image processing message: %v", err)
    } else {
        log.Printf("SUCCESS: Published image processing message for product ID: %d", product.ID)
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}
func GetProductsByUserHandler(c *gin.Context) {
    userID := 1
    if userIDQuery := c.DefaultQuery("user_id", ""); userIDQuery != "" {
        var err error
        userID, err = strconv.Atoi(userIDQuery)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
            return
        }
    }

    minPrice, _ := strconv.ParseFloat(c.DefaultQuery("min_price", "0"), 64)
    maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("max_price", "0"), 64)
    productName := c.DefaultQuery("product_name", "")


    products, err := models.GetProductsByUserID(userID, minPrice, maxPrice, productName)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, products)
}
