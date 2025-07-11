package api

import (
	"net/http"
	. "API-SUNAT2/service"
	. "API-SUNAT2/util"
	"github.com/gin-gonic/gin"
)

func setupRoutes(controller *UBLController) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(RequestIDMiddleware())

	router.GET("/health", controller.HealthCheck)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api/v1")
	{
		api.POST("/convert", controller.ConvertDocument)
		api.POST("/validate", controller.ValidateDocument)
		api.GET("/status/:correlationId", controller.GetDocumentStatus)
		api.GET("/xml/:filename", controller.GetXMLContent)
	}

	return router
}

// NewRouter crea y configura el router principal de la aplicación
func NewRouter() *gin.Engine {
	// Crear directorio para almacenar XML si no existe
	xmlStorePath := "./xml_output"
	
	// Crear servicios
	service := NewUBLConverterService(xmlStorePath)
	controller := NewUBLController(service)
	
	return setupRoutes(controller)
} 