package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "API-SUNAT2/model"
	. "API-SUNAT2/util"
	"github.com/sirupsen/logrus"
)

type UBLConverterService struct {
	validator    *ValidationService
	converter    *UBLConverter
	signer       *DigitalSignatureService
	logService   *LogService
	xmlStorePath string
}

// GetValidator retorna el validador para uso externo
func (s *UBLConverterService) GetValidator() *ValidationService {
	return s.validator
}

// GetXMLStorePath retorna la ruta de almacenamiento XML
func (s *UBLConverterService) GetXMLStorePath() string {
	return s.xmlStorePath
}

func NewUBLConverterService(xmlStorePath string) *UBLConverterService {
	logService := NewLogService()
	return &UBLConverterService{
		validator:    NewValidationService(logService.GetLogger()),
		converter:    NewUBLConverter(logService.GetLogger()),
		signer:       NewDigitalSignatureService(logService.GetLogger()),
		logService:   logService,
		xmlStorePath: xmlStorePath,
	}
}

func (s *UBLConverterService) ProcessDocument(doc *BusinessDocument, certPEM, keyPEM []byte) (*APIResponse, error) {
	startTime := time.Now()
	correlationID := GenerateCorrelationID()

	s.logService.LogInfo(correlationID, "PROCESS_DOCUMENT", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "Iniciando procesamiento de documento")

	validationErrors := s.validator.ValidateBusinessDocument(doc)
	if len(validationErrors) > 0 {
		s.logService.LogError(correlationID, "VALIDATION_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "VALIDATION_FAILED", "Documento no válido")
		return &APIResponse{
			Status:           "ERROR",
			CorrelationID:    correlationID,
			ProcessedAt:      time.Now(),
			ErrorCode:        "VALIDATION_FAILED",
			ErrorMessage:     "Documento no válido",
			ValidationErrors: validationErrors,
		}, nil
	}

	xmlData, err := s.converter.ConvertToUBL(doc)
	if err != nil {
		s.logService.LogError(correlationID, "CONVERSION_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "CONVERSION_FAILED", err.Error())
		return &APIResponse{
			Status:        "ERROR",
			CorrelationID: correlationID,
			ProcessedAt:   time.Now(),
			ErrorCode:     "CONVERSION_FAILED",
			ErrorMessage:  fmt.Sprintf("Error en conversión UBL: %v", err),
		}, nil
	}

	signedXML, err := s.signer.SignXML(xmlData, certPEM, keyPEM)
	if err != nil {
		s.logService.LogError(correlationID, "DIGITAL_SIGNATURE_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "SIGNATURE_FAILED", err.Error())
		return &APIResponse{
			Status:        "ERROR",
			CorrelationID: correlationID,
			ProcessedAt:   time.Now(),
			ErrorCode:     "SIGNATURE_FAILED",
			ErrorMessage:  fmt.Sprintf("Error en firma digital: %v", err),
		}, nil
	}

	fileName := fmt.Sprintf("%s-%s-%s-%s.xml", doc.Issuer.DocumentID, doc.Type, doc.Series, doc.Number)
	filePath := filepath.Join(s.xmlStorePath, fileName)

	err = os.WriteFile(filePath, signedXML, 0644)
	if err != nil {
		s.logService.LogError(correlationID, "FILE_SAVE_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "SAVE_FAILED", err.Error())
		return &APIResponse{
			Status:        "ERROR",
			CorrelationID: correlationID,
			ProcessedAt:   time.Now(),
			ErrorCode:     "SAVE_FAILED",
			ErrorMessage:  fmt.Sprintf("Error al guardar archivo: %v", err),
		}, nil
	}

	zipPath, err := ZipXMLFile(filePath)
	if err != nil {
		s.logService.LogError(correlationID, "ZIP_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "ZIP_FAILED", err.Error())
		return &APIResponse{
			Status:        "ERROR",
			CorrelationID: correlationID,
			ProcessedAt:   time.Now(),
			ErrorCode:     "ZIP_FAILED",
			ErrorMessage:  fmt.Sprintf("Error al crear ZIP: %v", err),
		}, nil
	}

	hash := sha256.Sum256(signedXML)
	xmlHash := hex.EncodeToString(hash[:])

	duration := time.Since(startTime).Milliseconds()

	s.logService.LogInfo(correlationID, "PROCESS_SUCCESS", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "Documento procesado exitosamente")

	response := &APIResponse{
		Status:        "SUCCESS",
		CorrelationID: correlationID,
		DocumentID:    fmt.Sprintf("%s-%s-%s-%s", doc.Issuer.DocumentID, doc.Type, doc.Series, doc.Number),
		XMLPath:       zipPath,
		XMLHash:       xmlHash,
		ProcessedAt:   time.Now(),
		Duration:      duration,
		Data: map[string]interface{}{
			"fileName": fileName,
			"fileSize": len(signedXML),
			"zipSize":  getFileSize(zipPath),
		},
		Message: fmt.Sprintf("El archivo ZIP fue generado exitosamente en: %s", zipPath),
	}

	return response, nil
}

func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

type UBLConverter struct {
	logger *logrus.Logger
}

func NewUBLConverter(logger *logrus.Logger) *UBLConverter {
	return &UBLConverter{logger: logger}
}

func (c *UBLConverter) ConvertToUBL(doc *BusinessDocument) ([]byte, error) {
	switch doc.Type {
	case "01", "03": // Factura o Boleta
		return c.convertToInvoice(doc)
	case "07": // Nota de Crédito
		return c.convertToCreditNote(doc)
	case "08": // Nota de Débito
		return c.convertToDebitNote(doc)
	default:
		return nil, fmt.Errorf("unsupported document type: %s", doc.Type)
	}
}

// --- CÓDIGO CORRECTO Y FINAL PARA LA FUNCIÓN convertToInvoice ---

func (c *UBLConverter) convertToInvoice(doc *BusinessDocument) ([]byte, error) {
	
	// Paso 1: Crear la instancia del struct UBLInvoice.
	// Esta estructura ya está definida correctamente en model/comprobante.go
	invoice := &UBLInvoice{

		// Paso 2: Asignar los valores a los campos de namespace.
		// Esto le dice a `xml.MarshalIndent` que cree la cabecera del XML correctamente.
		Xmlns:    "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		XmlnsCac: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XmlnsDs:  "http://www.w3.org/2000/09/xmldsig#",
		XmlnsExt: "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",

		// Paso 3: Rellenar los datos de la factura.
		// La estructura aquí coincide con la jerarquía correcta del XML.
		// UBLExtensions se deja como nil para que el servicio de firma digital lo cree
		UBLExtensions: nil,
		UBLVersionID:    "2.1",
		CustomizationID: "2.0",
		ProfileID:       "0101",
		ID:              fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		IssueDate:       doc.IssueDate,
		IssueTime:       time.Now().Format("15:04:05"),
		DueDate:         doc.DueDate,
		InvoiceTypeCode: UBLTypeCode{Value: doc.Type},
		DocumentCurrencyCode: UBLCurrencyCode{Value: doc.Currency},
		LineCountNumeric: len(doc.Items),
		Signature: &UBLSignature{
			ID: fmt.Sprintf("%s-%s", doc.Series, doc.Number),
			SignatoryParty: UBLSignatoryParty{
				PartyIdentification: UBLPartyIdentification{ID: UBLIDWithScheme{Value: doc.Issuer.DocumentID}},
				PartyName:           UBLPartyName{Name: doc.Issuer.Name},
			},
			DigitalSignatureAttachment: UBLDigitalSignatureAttachment{
				ExternalReference: UBLExternalReference{URI: "#SignatureSP"},
			},
		},
		AccountingSupplierParty: c.convertParty(doc.Issuer),
		AccountingCustomerParty: c.convertParty(doc.Customer),
		TaxTotal:                c.convertTaxTotals(doc.Taxes, doc.Currency),
		LegalMonetaryTotal:      c.convertLegalMonetaryTotal(doc.Totals, doc.Currency),
		InvoiceLines:            c.convertInvoiceLines(doc.Items, doc.Currency),
	}

	// Paso 4: Generar el XML.
	// `xml.MarshalIndent` ahora hará TODO el trabajo correctamente gracias a los pasos 1 y 2.
	// Ya no hay manipulación de strings.
	xmlData, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error al serializar UBLInvoice: %v", err)
	}

	// Paso 5: Añadir la declaración <?xml ...?> y devolver el resultado.
	xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	return append(xmlDeclaration, xmlData...), nil
}

// NOTA: Las funciones convertToCreditNote y convertToDebitNote no han sido modificadas.
// Deberás aplicar la misma lógica que en convertToInvoice si las necesitas.
func (c *UBLConverter) convertToCreditNote(doc *BusinessDocument) ([]byte, error) {
	// ... (código original sin modificar) ...
	return nil, fmt.Errorf("convertToCreditNote no implementado con las nuevas correcciones")
}

func (c *UBLConverter) convertToDebitNote(doc *BusinessDocument) ([]byte, error) {
	// ... (código original sin modificar) ...
	return nil, fmt.Errorf("convertToDebitNote no implementado con las nuevas correcciones")
}


func (c *UBLConverter) convertParty(party Party) *UBLParty {
	schemeID := "6" // RUC por defecto
	if party.DocumentType == "1" {
		schemeID = "1" // DNI
	}

	return &UBLParty{
		Party: UBLPartyDetail{
			PartyIdentification: UBLPartyIdentification{
				ID: UBLIDWithScheme{
					SchemeID:         schemeID,
					SchemeName:       "SUNAT:Identificador de Documento de Identidad",
					SchemeAgencyName: "PE:SUNAT",
					Value:            party.DocumentID,
				},
			},
			PartyName: []UBLPartyName{
				{
					Name: party.Name,
				},
			},
			PartyLegalEntity: []UBLPartyLegalEntity{
				{
					RegistrationName: party.Name,
					RegistrationAddress: UBLRegistrationAddress{
						ID: UBLIDWithScheme{
							SchemeAgencyName: "PE:INEI",
							SchemeName:       "Ubigeos",
							Value:            "140101", // OJO: Ubigeo hardcodeado
						},
						CityName:         party.Address.City,
						CountrySubentity: party.Address.Province,
						District:         party.Address.District,
						AddressLine: UBLAddressLine{
							Line: fmt.Sprintf("%s - %s - %s - %s", party.Address.Street, party.Address.District, party.Address.Province, party.Address.Department),
						},
						Country: UBLCountry{
							IdentificationCode: UBLIDWithScheme{
								Value: party.Address.Country,
							},
						},
					},
				},
			},
		},
	}
}

func (c *UBLConverter) convertTaxTotals(taxes []TaxTotal, currency string) []UBLTaxTotal {
	var taxTotals []UBLTaxTotal
	if len(taxes) == 0 {
		return nil
	}
	
	// Agrupa todos los subtotales en un único TaxTotal global
	var globalSubtotals []UBLTaxSubtotal
	for _, tax := range taxes {
		globalSubtotals = append(globalSubtotals, UBLTaxSubtotal{
			TaxableAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      Decimal2(tax.TaxBase),
			},
			TaxAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      Decimal2(tax.TaxAmount),
			},
			TaxCategory: UBLTaxCategory{
				ID: UBLIDWithScheme{
					Value: "S",
				},
				Percent: tax.TaxRate,
				TaxScheme: UBLTaxScheme{
					ID:          UBLIDWithScheme{Value: "1000"}, // IGV
					Name:        "IGV",
					TaxTypeCode: "VAT",
				},
			},
		})
	}

	taxTotals = append(taxTotals, UBLTaxTotal{
		TaxAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      Decimal2(taxes[0].TaxAmount), // Asume que solo hay un tipo de impuesto global
		},
		TaxSubtotals: globalSubtotals,
	})

	return taxTotals
}

func (c *UBLConverter) getTaxName(taxType string) string {
	switch taxType {
	case "1000":
		return "IGV"
	case "2000":
		return "ISC"
	case "7152":
		return "ICBPER"
	default:
		return "TAX"
	}
}

func (c *UBLConverter) convertLegalMonetaryTotal(totals DocumentTotals, currency string) UBLLegalMonetaryTotal {
	return UBLLegalMonetaryTotal{
		LineExtensionAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      Decimal2(totals.SubTotal),
		},
		TaxInclusiveAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      Decimal2(totals.TotalAmount),
		},
		PayableAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      Decimal2(totals.PayableAmount),
		},
	}
}

func (c *UBLConverter) convertInvoiceLines(items []DocumentItem, currency string) []UBLInvoiceLine {
	var lines []UBLInvoiceLine
	for i, item := range items {
		line := UBLInvoiceLine{
			ID: fmt.Sprintf("%d", i+1),
			InvoicedQuantity: UBLQuantityWithUnit{
				UnitCode: item.UnitCode,
				Value:    Decimal2(item.Quantity),
			},
			LineExtensionAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      Decimal2(item.LineTotal),
			},
			PricingReference: &UBLPricingReference{
				AlternativeConditionPrice: UBLAlternativeConditionPrice{
					PriceAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      Decimal2(item.UnitPrice * (1 + (item.Taxes[0].TaxRate / 100))), // Precio de venta unitario CON IGV
					},
					PriceTypeCode: UBLIDWithScheme{
						Value: "01", // Precio de Venta Unitario (incluye impuestos)
					},
				},
			},
			TaxTotal: c.convertItemTaxes(item.Taxes, currency),
			Item: UBLItem{
				Description: item.Description,
				SellersItemIdentification: &UBLSellersItemIdentification{
					ID: item.ID,
				},
			},
			Price: UBLPrice{
				PriceAmount: UBLAmountWithCurrency{
					CurrencyID: currency,
					Value:      Decimal2(item.UnitPrice), // Valor unitario SIN IGV
				},
			},
		}
		lines = append(lines, line)
	}
	return lines
}

func (c *UBLConverter) convertItemTaxes(taxes []Tax, currency string) []UBLTaxTotal {
	var taxTotals []UBLTaxTotal
	for _, tax := range taxes {
		taxTotal := UBLTaxTotal{
			TaxAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      Decimal2(tax.TaxAmount),
			},
			TaxSubtotals: []UBLTaxSubtotal{
				{
					TaxableAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      Decimal2(tax.TaxBase),
					},
					TaxAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      Decimal2(tax.TaxAmount),
					},
					TaxCategory: UBLTaxCategory{
						ID: UBLIDWithScheme{
							Value: "S", // Impuesto que afecta a la base imponible
						},
						Percent: tax.TaxRate,
						TaxExemptionReasonCode: UBLIDWithScheme{
							Value: "10", // Gravado - Operación Onerosa
						},
						TaxScheme: UBLTaxScheme{
							ID:          UBLIDWithScheme{Value: tax.TaxType},
							Name:        c.getTaxName(tax.TaxType),
							TaxTypeCode: "VAT",
						},
					},
				},
			},
		}
		taxTotals = append(taxTotals, taxTotal)
	}
	return taxTotals
}