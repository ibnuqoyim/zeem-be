package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"zeem/internal/config"
	"zeem/internal/handlers"
	"zeem/internal/services"
	"zeem/internal/static"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Starting application with config: Environment=%s, Port=%s\n", cfg.Environment, cfg.Port)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Log allowed origins
	log.Printf("Allowed origins: %v\n", cfg.AllowedOrigins)

	// Initialize services
	roomManager := services.NewRoomManager()
	webrtcManager := services.NewWebRTCManager()
	wsHandler := handlers.NewWebSocketHandler(roomManager, webrtcManager)

	router := gin.Default()

	// Recovery middleware with logger
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// CORS policy based on configuration
		if cfg.Environment == "production" {
			// In production, check against allowed origins
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			if allowed {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		} else {
			// In development, allow all
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

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
	staticFiles, err := fs.Sub(static.StaticFiles, "client")
	if err != nil {
		log.Fatal("Failed to setup static files:", err)
	}

	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Writer.Header().Add("Cache-Control", "no-cache")
			http.FileServer(http.FS(staticFiles)).ServeHTTP(c.Writer, c.Request)
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

	serverAddr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting HTTP server on %s in %s mode\n", serverAddr, cfg.Environment)
	if err := router.Run(serverAddr); err != nil {
		log.Fatal("Failed to start HTTP server: ", err)
	}
}
