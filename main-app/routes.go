// cmd/routes.go
package main

import (
	"github.com/gin-gonic/gin"
	"design-carousel-service/api"
)

func initializeRoutes(router *gin.Engine) {
	log.Info("Initializing DesignCarousel API routes...")

	router.GET("/isalive", api.IsAlive)
	router.GET("/carousel", api.ListSlides)
	router.POST("/carousel", api.AddSlide)
	router.DELETE("/carousel/:id", api.DeleteSlide)
	router.GET("/serve", api.ServeImage)
}

