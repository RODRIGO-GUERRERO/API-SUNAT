package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	. "API-SUNAT2/model"
	. "API-SUNAT2/service"
	"github.com/gin-gonic/gin"
)

type UBLController struct {
	service *UBLConverterService
}

func NewUBLController(service *UBLConverterService) *UBLController {
	return &UBLController{service: service}
}

func (ctrl *UBLController) ConvertDocument(c *gin.Context) {
	var request struct {
		Document    BusinessDocument `json:"document"`
		Certificate string          `json:"certificate"`
		PrivateKey  string          `json:"privateKey"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_INVALID_REQUEST",
			ErrorMessage: fmt.Sprintf("Invalid request format: %v", err),
			ProcessedAt:  time.Now(),
		})
		return
	}

	certPEM, err := base64.StdEncoding.DecodeString(request.Certificate)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_INVALID_CERTIFICATE",
			ErrorMessage: "Invalid certificate format",
			ProcessedAt:  time.Now(),
		})
		return
	}

	keyPEM, err := base64.StdEncoding.DecodeString(request.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_INVALID_PRIVATE_KEY",
			ErrorMessage: "Invalid private key format",
			ProcessedAt:  time.Now(),
		})
		return
	}

	response, err := ctrl.service.ProcessDocument(&request.Document, certPEM, keyPEM)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_PROCESSING_FAILED",
			ErrorMessage: fmt.Sprintf("Processing failed: %v", err),
			ProcessedAt:  time.Now(),
		})
		return
	}

	statusCode := http.StatusOK
	if response.Status == "error" {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, response)
}

func (ctrl *UBLController) GetDocumentStatus(c *gin.Context) {
	correlationID := c.Param("correlationId")
	c.JSON(http.StatusOK, APIResponse{
		Status:        "success",
		CorrelationID: correlationID,
		ProcessedAt:   time.Now(),
		Data: map[string]interface{}{
			"message": "Document processing completed successfully",
		},
	})
}

func (ctrl *UBLController) ValidateDocument(c *gin.Context) {
	var doc BusinessDocument

	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_INVALID_REQUEST",
			ErrorMessage: fmt.Sprintf("Invalid request format: %v", err),
			ProcessedAt:  time.Now(),
		})
		return
	}

	validationErrors := ctrl.service.GetValidator().ValidateBusinessDocument(&doc)

	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:           "error",
			ErrorCode:        "ERR_VALIDATION_FAILED",
			ErrorMessage:     "Document validation failed",
			ValidationErrors: validationErrors,
			ProcessedAt:      time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Status:      "success",
		ProcessedAt: time.Now(),
		Data: map[string]interface{}{
			"message": "Document validation passed",
		},
	})
}

func (ctrl *UBLController) GetXMLContent(c *gin.Context) {
	filename := c.Param("filename")

	if !strings.HasSuffix(filename, ".xml") {
		c.JSON(http.StatusBadRequest, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_INVALID_FILENAME",
			ErrorMessage: "Invalid filename format",
			ProcessedAt:  time.Now(),
		})
		return
	}

	filePath := fmt.Sprintf("%s/%s", ctrl.service.GetXMLStorePath(), filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Status:       "error",
			ErrorCode:    "ERR_FILE_NOT_FOUND",
			ErrorMessage: "XML file not found",
			ProcessedAt:  time.Now(),
		})
		return
	}

	c.Header("Content-Type", "application/xml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/xml", content)
}

func (ctrl *UBLController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"service":   "UBL Converter API",
	})
}

// Handlers antiguos para compatibilidad
func ValidateHandler(c *gin.Context) {}
func ConvertHandler(c *gin.Context) {}
func GetXMLHandler(c *gin.Context) {}
func HealthHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
} 