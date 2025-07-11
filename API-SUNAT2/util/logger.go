package util

import (
	"github.com/sirupsen/logrus"
)

// Definir OperationLog localmente si no está disponible en el modelo
type OperationLog struct {
	Timestamp     string
	CorrelationID string
	Level         string
	Operation     string
	DocumentType  string
	DocumentID    string
	Duration      int64
	Status        string
	Error         string
	ErrorCode     string
	StackTrace    string
}

type LogService struct {
	logger *logrus.Logger
}

func NewLogService() *LogService {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	return &LogService{logger: logger}
}

func (l *LogService) LogOperation(log OperationLog) {
	l.logger.WithFields(logrus.Fields{
		"correlationId": log.CorrelationID,
		"operation":     log.Operation,
		"documentType":  log.DocumentType,
		"documentId":    log.DocumentID,
		"duration":      log.Duration,
		"status":        log.Status,
		"error":         log.Error,
		"errorCode":     log.ErrorCode,
		"timestamp":     log.Timestamp,
		"level":         log.Level,
		"stackTrace":    log.StackTrace,
	}).Info("Operation logged")
}

func (l *LogService) LogError(correlationID, operation, documentType, documentID, errorCode, errorMsg string) {
	l.logger.WithFields(logrus.Fields{
		"correlationId": correlationID,
		"operation":     operation,
		"documentType":  documentType,
		"documentId":    documentID,
		"errorCode":     errorCode,
	}).Error(errorMsg)
}

func (l *LogService) LogInfo(correlationID, operation, documentType, documentID, message string) {
	l.logger.WithFields(logrus.Fields{
		"correlationId": correlationID,
		"operation":     operation,
		"documentType":  documentType,
		"documentId":    documentID,
	}).Info(message)
}

// GetLogger retorna el logger interno para uso en otros servicios
func (l *LogService) GetLogger() *logrus.Logger {
	return l.logger
} 