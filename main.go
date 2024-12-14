package main

import (
	"AsyncProd/config"
	"AsyncProd/handlers"
	"AsyncProd/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	
	gin.SetMode(gin.ReleaseMode)

	// init all the imp stuff (DB, Redis, RabbitMQ, S3).
	config.InitDB()
	defer config.CloseDB()
	config.InitRedis()
	defer config.CloseRedis()
	config.InitRabbitMQ()
	defer config.CloseRabbitMQ()
	config.InitS3()

	// Start the image processing service in the background.
	go func() {
		log.Println("Starting image processing service.")
		services.ProcessImageFromQueue()
	}()

	
	r := gin.New()

	// Add middlewares for logging and handling crashes.
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	
	v1 := r.Group("/api/v1")
	{
		v1.POST("/products", handlers.CreateProductHandler)
		v1.GET("/products/:id", handlers.GetProductByIDHandler)
		v1.GET("/products", handlers.GetProductsByUserHandler)
		v1.PUT("/products", handlers.UpdateProductHandler)
	}

	
	r.GET("/health", healthCheckHandler)
	r.GET("/redis-health", redisHealthCheckHandler)

	// to avoid bad actors we kinda limit
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// we start in a separate goroutine.
	go func() {
		log.Printf("Server running at %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for a signal (like CTRL+C) to shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server.")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down: %v", err)
	}

	log.Println("Server stopped.")
}

// db health 
func healthCheckHandler(c *gin.Context) {
	if err := config.DB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}

// redis health
func redisHealthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := config.RedisClient.Ping(ctx).Result(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"redis":  "connected",
	})
}
