package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"zeem/internal/handlers"
	"zeem/internal/services"
)

func main() {
	// Initialize services
	roomManager := services.NewRoomManager()
	webrtcManager := services.NewWebRTCManager()
	wsHandler := handlers.NewWebSocketHandler(roomManager, webrtcManager)

	router := gin.Default()

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		// Strict CORS policy
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://"+c.Request.Host)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Security headers
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; connect-src 'self' wss: https: ws:; media-src 'self' mediastream: blob:; img-src 'self' data: blob:; worker-src 'self'; child-src 'self'")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Writer.Header().Add("Cache-Control", "no-cache")
			http.FileServer(http.Dir("client")).ServeHTTP(c.Writer, c.Request)
			return
		}
		c.Next()
	})

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleConnection)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Start HTTPS server
	log.Println("Starting HTTPS server on :80")
	if err := router.RunTLS(":80", "certs/cert.pem", "certs/key.pem"); err != nil {
		log.Fatal("Failed to start HTTPS server: ", err)
	}
}
