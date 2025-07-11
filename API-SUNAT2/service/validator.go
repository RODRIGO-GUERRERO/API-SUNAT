package service

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	. "API-SUNAT2/model"
	"github.com/sirupsen/logrus"
)

type ValidationService struct {
	logger *logrus.Logger
}

func NewValidationService(logger *logrus.Logger) *ValidationService {
	return &ValidationService{logger: logger}
}

func (v *ValidationService) ValidateBusinessDocument(doc *BusinessDocument) []ValidationError {
	var errors []ValidationError

	// Validar RUC
	if !v.isValidRUC(doc.Issuer.DocumentID) {
		errors = append(errors, ValidationError{
			Field:    "issuer.documentId",
			Expected: "Valid RUC format",
			Received: doc.Issuer.DocumentID,
			Rule:     "ruc_validation",
			Message:  "RUC format is invalid",
		})
	}

	// Validar tipo de documento
	if !v.isValidDocumentType(doc.Type) {
		errors = append(errors, ValidationError{
			Field:    "type",
			Expected: "Valid document type (01, 03, 07, 08)",
			Received: doc.Type,
			Rule:     "document_type_validation",
			Message:  "Document type is not valid",
		})
	}

	// Validar moneda
	if !v.isValidCurrency(doc.Currency) {
		errors = append(errors, ValidationError{
			Field:    "currency",
			Expected: "Valid currency code (PEN, USD, EUR)",
			Received: doc.Currency,
			Rule:     "currency_validation",
			Message:  "Currency code is not valid",
		})
	}

	// Validar totales
	calculatedTotal := v.calculateTotal(doc)
	if calculatedTotal != doc.Totals.TotalAmount {
		errors = append(errors, ValidationError{
			Field:    "totals.totalAmount",
			Expected: fmt.Sprintf("%.2f", calculatedTotal),
			Received: fmt.Sprintf("%.2f", doc.Totals.TotalAmount),
			Rule:     "sum_validation",
			Message:  "Total amount calculation mismatch",
		})
	}

	// Validar fecha
	if !v.isValidDate(doc.IssueDate) {
		errors = append(errors, ValidationError{
			Field:    "issueDate",
			Expected: "Valid date format YYYY-MM-DD",
			Received: doc.IssueDate,
			Rule:     "date_validation",
			Message:  "Issue date format is invalid",
		})
	}

	// Validar items
	for i, item := range doc.Items {
		if item.Quantity <= 0 {
			errors = append(errors, ValidationError{
				Field:    fmt.Sprintf("items[%d].quantity", i),
				Expected: "Greater than 0",
				Received: fmt.Sprintf("%.2f", item.Quantity),
				Rule:     "quantity_validation",
				Message:  "Quantity must be greater than 0",
			})
		}

		if item.UnitPrice <= 0 {
			errors = append(errors, ValidationError{
				Field:    fmt.Sprintf("items[%d].unitPrice", i),
				Expected: "Greater than 0",
				Received: fmt.Sprintf("%.2f", item.UnitPrice),
				Rule:     "price_validation",
				Message:  "Unit price must be greater than 0",
			})
		}
	}

	return errors
}

func (v *ValidationService) isValidRUC(ruc string) bool {
	if len(ruc) != 11 {
		return false
	}
	if matched, _ := regexp.MatchString(`^\d{11}$`, ruc); !matched {
		return false
	}
	weights := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(ruc[i]))
		sum += digit * weights[i]
	}
	remainder := sum % 11
	checkDigit := 11 - remainder
	if checkDigit == 11 {
		checkDigit = 0
	} else if checkDigit == 10 {
		checkDigit = 1
	}
	lastDigit, _ := strconv.Atoi(string(ruc[10]))
	return checkDigit == lastDigit
}

func (v *ValidationService) isValidDocumentType(docType string) bool {
	validTypes := map[string]bool{
		"01": true,
		"03": true,
		"07": true,
		"08": true,
	}
	return validTypes[docType]
}

func (v *ValidationService) isValidCurrency(currency string) bool {
	validCurrencies := map[string]bool{
		"PEN": true,
		"USD": true,
		"EUR": true,
	}
	return validCurrencies[currency]
}

func (v *ValidationService) isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func (v *ValidationService) calculateTotal(doc *BusinessDocument) float64 {
	total := doc.Totals.SubTotal
	for _, tax := range doc.Taxes {
		total += tax.TaxAmount
	}
	return total
} 