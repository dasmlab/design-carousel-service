package main

import (
	"os"
//	"time"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
	"github.com/Depado/ginprom"
	"github.com/gin-contrib/cors"

	"design-carousel-service/logutil"
	"design-carousel-service/docs"
//	"design-carousel-service/api"
)

// VARS
const version = "0.0.1"
var componentName = "design-carousel-main"
var log = logutil.InitLogger(componentName)

// @title DASMLAB DesignCarousel Service
// @version 0.0.1
// @description APIs for Image Carousel Management
// @BasePath /

func main() {
	log.Infof("DASMLAB DesignCarousel Service - Starting %s", componentName)
	docs.SwaggerInfo.Version = version

	// Set Gin Prod mode
	gin.SetMode(gin.ReleaseMode)

	// (Optional) Check ENV vars, such as storage backend, here if needed
	storageBackend := os.Getenv("STORAGE_BACKEND")
	if storageBackend == "" {
		storageBackend = "memory"
	}
	log.Infof("Using storage backend: %s", storageBackend)

	// Main Router (API)
	mainRouter := gin.Default()
	mainRouter.Use(cors.Default()) // Allow all CORS for now

	// Metrics Router (Prometheus, /metrics)
	metricsRouter := gin.Default()
	p := ginprom.New(
		ginprom.Engine(metricsRouter),
		ginprom.Subsystem("gin"),
		ginprom.Path("/metrics"),
	)
	mainRouter.Use(p.Instrument())

	// Add Swagger UI
	mainRouter.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize API Routes (see routes.go)
	initializeRoutes(mainRouter)

	// Launch Metrics Server (out-of-band)
	go func() {
		log.Infof("Starting metrics server on :9222")
		if err := metricsRouter.Run(":9222"); err != nil {
			log.Fatalf("Metrics Server Error: %v", err)
		}
	}()

	// Launch Main API Server
	log.Info("Start main API Server listening on :10022")
	if err := mainRouter.Run(":10022"); err != nil {
		log.Fatalf("Main Server Error: %v", err)
	}
}

