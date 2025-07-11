package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "API-SUNAT2/model"
	. "API-SUNAT2/util"
	"github.com/sirupsen/logrus"
)

type UBLConverterService struct {
	validator     *ValidationService
	converter     *UBLConverter
	signer        *DigitalSignatureService
	logService    *LogService
	xmlStorePath  string
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
		validator:     NewValidationService(logService.GetLogger()),
		converter:     NewUBLConverter(logService.GetLogger()),
		signer:        NewDigitalSignatureService(logService.GetLogger()),
		logService:    logService,
		xmlStorePath:  xmlStorePath,
	}
}

func (s *UBLConverterService) ProcessDocument(doc *BusinessDocument, certPEM, keyPEM []byte) (*APIResponse, error) {
	startTime := time.Now()
	correlationID := GenerateCorrelationID()

	// Log inicio del proceso
	s.logService.LogInfo(correlationID, "PROCESS_DOCUMENT", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "Iniciando procesamiento de documento")

	// Validar documento
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

	// Convertir a UBL
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

	// Agregar firma UBL al XML
	xmlData, err = s.addUBLSignature(xmlData, doc)
	if err != nil {
		s.logService.LogError(correlationID, "UBL_SIGNATURE_ERROR", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "UBL_SIGNATURE_FAILED", err.Error())
		return &APIResponse{
			Status:        "ERROR",
			CorrelationID: correlationID,
			ProcessedAt:   time.Now(),
			ErrorCode:     "UBL_SIGNATURE_FAILED",
			ErrorMessage:  fmt.Sprintf("Error al agregar firma UBL: %v", err),
		}, nil
	}

	// Firmar digitalmente
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

	// Generar nombre de archivo
	fileName := fmt.Sprintf("%s-%s-%s-%s.xml", doc.Issuer.DocumentID, doc.Type, doc.Series, doc.Number)
	filePath := filepath.Join(s.xmlStorePath, fileName)

	// Guardar XML firmado
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

	// Crear archivo ZIP
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

	// Calcular hash del XML
	hash := sha256.Sum256(signedXML)
	xmlHash := hex.EncodeToString(hash[:])

	// Calcular duración
	duration := time.Since(startTime).Milliseconds()

	// Log éxito
	s.logService.LogInfo(correlationID, "PROCESS_SUCCESS", doc.Type, fmt.Sprintf("%s-%s", doc.Series, doc.Number), "Documento procesado exitosamente")

	// Retornar respuesta exitosa
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
		Message:       fmt.Sprintf("El archivo ZIP fue generado exitosamente en: %s", zipPath),
	}

	return response, nil
}

func (s *UBLConverterService) addUBLSignature(xmlData []byte, doc *BusinessDocument) ([]byte, error) {
	// Crear firma UBL
	ublSignature := &UBLSignature{
		ID: fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		SignatoryParty: UBLSignatoryParty{
			PartyIdentification: UBLPartyIdentification{
				ID: UBLIDWithScheme{
					Value: doc.Issuer.DocumentID,
				},
			},
			PartyName: UBLPartyName{
				Name: doc.Issuer.Name,
			},
		},
		DigitalSignatureAttachment: UBLDigitalSignatureAttachment{
			ExternalReference: UBLExternalReference{
				URI: "#SignatureSP",
			},
		},
	}

	// Convertir XML a string para manipulación
	xmlStr := string(xmlData)

	// Buscar la posición después de UBLVersionID para insertar la firma
	ublVersionEnd := strings.Index(xmlStr, "</cbc:UBLVersionID>")
	if ublVersionEnd == -1 {
		return nil, fmt.Errorf("UBLVersionID not found")
	}

	// Crear XML de la firma UBL
	signatureXML, err := xml.MarshalIndent(ublSignature, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UBL signature: %v", err)
	}

	// Insertar la firma después de UBLVersionID
	xmlStr = xmlStr[:ublVersionEnd+len("</cbc:UBLVersionID>")] + "\n" + string(signatureXML) + xmlStr[ublVersionEnd+len("</cbc:UBLVersionID>"):]

	return []byte(xmlStr), nil
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

func (c *UBLConverter) convertToInvoice(doc *BusinessDocument) ([]byte, error) {
	invoice := &UBLInvoice{
		XMLName:       xml.Name{Local: "Invoice"},
		Xmlns:         "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		UBLExtensions: &UBLExtensions{
			UBLExtension: UBLExtension{
				ExtensionContent: ExtensionContent{
					Signature: XMLSignature{}, // Se reemplazará luego por la firma real
				},
			},
		},
		UBLVersionID: "2.1",
		CustomizationID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			Value:            "2.0",
		},
		ProfileID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			SchemeName:       "Tipo de Operacion",
			SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo51",
			Value:            "0101",
		},
		ID: fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		IssueDate: doc.IssueDate,
		IssueTime: "10:30:00",
		DueDate:   doc.IssueDate,
		InvoiceTypeCode: UBLTypeCode{
			ListAgencyName: "PE:SUNAT",
			ListID:         "0101",
			ListName:       "Tipo de Documento",
			ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo01",
			Name:           "Tipo de Operacion",
			Value:          doc.Type,
		},
		DocumentCurrencyCode: UBLIDWithScheme{
			SchemeAgencyName: "United Nations Economic Commission for Europe",
			SchemeID:         "ISO 4217 Alpha",
			SchemeName:       "Currency",
			Value:            doc.Currency,
		},
		LineCountNumeric:       len(doc.Items),
		Note:                   "",
		Signature:              c.createUBLSignature(doc),
		AccountingSupplierParty: c.convertParty(doc.Issuer),
		AccountingCustomerParty: c.convertParty(doc.Customer),
		PaymentTerms: []UBLPaymentTerms{
			{
				ID:             "FormaPago",
				PaymentMeansID: "Contado",
			},
		},
		TaxTotal:           c.convertTaxTotals(doc.Taxes, doc.Currency),
		LegalMonetaryTotal: c.convertLegalMonetaryTotal(doc.Totals, doc.Currency),
		InvoiceLines:       c.convertInvoiceLines(doc.Items, doc.Currency),
	}
	if doc.Type == "03" {
		invoice.Note = "TRANSFERENCIA GRATUITA DE UN BIEN Y/O SERVICIO PRESTADO GRATUITAMENTE"
	}
	xmlData, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling invoice XML: %v", err)
	}
	xmlDeclaration := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	return append(xmlDeclaration, xmlData...), nil
}

func (c *UBLConverter) convertToCreditNote(doc *BusinessDocument) ([]byte, error) {
	creditNote := &UBLCreditNote{
		Xmlns:                  "urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2",
		XmlnsCac:               "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc:               "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XmlnsDs:                "http://www.w3.org/2000/09/xmldsig#",
		XmlnsExt:               "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
		UBLExtensions:          nil,
		UBLVersionID:           "2.1",
		CustomizationID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			Value:            "2.0",
		},
		ProfileID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			SchemeName:       "Tipo de Operacion",
			SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo51",
			Value:            "0101",
		},
		ID:                     fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		IssueDate:              doc.IssueDate,
		IssueTime:              "10:30:00",
		CreditNoteTypeCode: UBLTypeCode{
			ListAgencyName: "PE:SUNAT",
			ListID:         "0101",
			ListName:       "Tipo de Documento",
			ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo09",
			Name:           "Tipo de Operacion",
			Value:          doc.Type,
		},
		DocumentCurrencyCode: UBLIDWithScheme{
			SchemeAgencyName: "United Nations Economic Commission for Europe",
			SchemeID:         "ISO 4217 Alpha",
			SchemeName:       "Currency",
			Value:            doc.Currency,
		},
		LineCountNumeric:       len(doc.Items),
		DiscrepancyResponse:    []UBLDiscrepancyResponse{},
		BillingReference:       []UBLBillingReference{},
		Signature:              nil,
		AccountingSupplierParty: c.convertParty(doc.Issuer),
		AccountingCustomerParty: c.convertParty(doc.Customer),
		PaymentTerms: []UBLPaymentTerms{
			{
				ID:             "FormaPago",
				PaymentMeansID: "Contado",
			},
		},
		TaxTotal:           c.convertTaxTotals(doc.Taxes, doc.Currency),
		LegalMonetaryTotal: c.convertLegalMonetaryTotal(doc.Totals, doc.Currency),
		CreditNoteLines:    c.convertCreditNoteLines(doc.Items, doc.Currency),
	}
	if doc.Reference != nil {
		creditNote.DiscrepancyResponse = []UBLDiscrepancyResponse{
			{
				ReferenceID:  doc.Reference.DocumentID,
				ResponseCode: "01",
				Description:  doc.Reference.Reason,
			},
		}
		creditNote.BillingReference = []UBLBillingReference{
			{
				InvoiceDocumentReference: UBLDocumentReference{
					ID:              doc.Reference.DocumentID,
					IssueDate:       doc.Reference.IssueDate,
					DocumentTypeCode: doc.Reference.DocumentType,
				},
			},
		}
	}
	xmlData, err := xml.MarshalIndent(creditNote, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling credit note XML: %v", err)
	}
	xmlDeclaration := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	return append(xmlDeclaration, xmlData...), nil
}

func (c *UBLConverter) convertToDebitNote(doc *BusinessDocument) ([]byte, error) {
	debitNote := &UBLDebitNote{
		Xmlns:                  "urn:oasis:names:specification:ubl:schema:xsd:DebitNote-2",
		XmlnsCac:               "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XmlnsCbc:               "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
		XmlnsDs:                "http://www.w3.org/2000/09/xmldsig#",
		XmlnsExt:               "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
		UBLExtensions:          nil,
		UBLVersionID:           "2.1",
		CustomizationID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			Value:            "2.0",
		},
		ProfileID: UBLIDWithScheme{
			SchemeAgencyName: "PE:SUNAT",
			SchemeName:       "Tipo de Operacion",
			SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo51",
			Value:            "0101",
		},
		ID:                     fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		IssueDate:              doc.IssueDate,
		IssueTime:              "10:30:00",
		DebitNoteTypeCode: UBLTypeCode{
			ListAgencyName: "PE:SUNAT",
			ListID:         "0101",
			ListName:       "Tipo de Documento",
			ListURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo10",
			Name:           "Tipo de Operacion",
			Value:          doc.Type,
		},
		DocumentCurrencyCode: UBLIDWithScheme{
			SchemeAgencyName: "United Nations Economic Commission for Europe",
			SchemeID:         "ISO 4217 Alpha",
			SchemeName:       "Currency",
			Value:            doc.Currency,
		},
		LineCountNumeric:       len(doc.Items),
		DiscrepancyResponse:    []UBLDiscrepancyResponse{},
		BillingReference:       []UBLBillingReference{},
		Signature:              nil,
		AccountingSupplierParty: c.convertParty(doc.Issuer),
		AccountingCustomerParty: c.convertParty(doc.Customer),
		PaymentTerms: []UBLPaymentTerms{
			{
				ID:             "FormaPago",
				PaymentMeansID: "Contado",
			},
		},
		TaxTotal:           c.convertTaxTotals(doc.Taxes, doc.Currency),
		LegalMonetaryTotal: c.convertLegalMonetaryTotal(doc.Totals, doc.Currency),
		DebitNoteLines:     c.convertDebitNoteLines(doc.Items, doc.Currency),
	}
	if doc.Reference != nil {
		debitNote.DiscrepancyResponse = []UBLDiscrepancyResponse{
			{
				ReferenceID:  doc.Reference.DocumentID,
				ResponseCode: "01",
				Description:  doc.Reference.Reason,
			},
		}
		debitNote.BillingReference = []UBLBillingReference{
			{
				InvoiceDocumentReference: UBLDocumentReference{
					ID:              doc.Reference.DocumentID,
					IssueDate:       doc.Reference.IssueDate,
					DocumentTypeCode: doc.Reference.DocumentType,
				},
			},
		}
	}
	xmlData, err := xml.MarshalIndent(debitNote, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling debit note XML: %v", err)
	}
	xmlDeclaration := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	return append(xmlDeclaration, xmlData...), nil
}

func (c *UBLConverter) createUBLSignature(doc *BusinessDocument) *UBLSignature {
	return &UBLSignature{
		ID: fmt.Sprintf("%s-%s", doc.Series, doc.Number),
		SignatoryParty: UBLSignatoryParty{
			PartyIdentification: UBLPartyIdentification{
				ID: UBLIDWithScheme{
					Value: doc.Issuer.DocumentID,
				},
			},
			PartyName: UBLPartyName{
				Name: doc.Issuer.Name,
			},
		},
		DigitalSignatureAttachment: UBLDigitalSignatureAttachment{
			ExternalReference: UBLExternalReference{
				URI: "#SignatureSP",
			},
		},
	}
}

func (c *UBLConverter) convertParty(party Party) UBLParty {
	schemeID := "6"
	if party.DocumentType == "1" {
		schemeID = "1"
	}
	return UBLParty{
		Party: UBLPartyDetail{
			PartyIdentification: []UBLPartyIdentification{
				{
					ID: UBLIDWithScheme{
						SchemeAgencyName: "PE:SUNAT",
						SchemeID:         schemeID,
						SchemeName:       "Documento de Identidad",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
						Value:            party.DocumentID,
					},
				},
			},
			PartyName: []UBLPartyName{
				{Name: party.Name},
			},
			RegistrationAddress: UBLRegistrationAddress{
				ID: UBLIDWithScheme{
					SchemeAgencyName: "PE:INEI",
					SchemeName:       "Ubigeos",
					Value:            "140101",
				},
				AddressTypeCode: UBLIDWithScheme{
					SchemeAgencyName: "PE:SUNAT",
					SchemeName:       "Establecimientos anexos",
					Value:          "0000",
				},
				CityName:         party.Address.City,
				CountrySubentity: party.Address.Province,
				District:         party.Address.District,
				AddressLine: UBLAddressLine{
					Line: fmt.Sprintf("%s - %s - %s - %s", party.Address.Street, party.Address.District, party.Address.Province, party.Address.Department),
				},
				Country: UBLCountry{
					IdentificationCode: UBLIDWithScheme{
						SchemeAgencyName: "United Nations Economic Commission for Europe",
						SchemeID:         "ISO 3166-1",
						SchemeName:       "Country",
						Value:          party.Address.Country,
					},
				},
			},
			PartyTaxScheme: []UBLPartyTaxScheme{
				{
					RegistrationName: party.Name,
					CompanyID: UBLIDWithScheme{
						SchemeAgencyName: "PE:SUNAT",
						SchemeID:         schemeID,
						SchemeName:       "SUNAT:Identificador de Documento de Identidad",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
						Value:            party.DocumentID,
					},
					TaxScheme: UBLTaxScheme{
						ID: UBLIDWithScheme{
							SchemeAgencyName: "PE:SUNAT",
							SchemeID:         schemeID,
							SchemeName:       "SUNAT:Identificador de Documento de Identidad",
							SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo06",
							Value:            party.DocumentID,
						},
					},
				},
			},
			PartyLegalEntity: []UBLPartyLegalEntity{
				{
					RegistrationName: party.Name,
					RegistrationAddress: UBLRegistrationAddress{
						ID: UBLIDWithScheme{
							SchemeAgencyName: "PE:INEI",
							SchemeName:       "Ubigeos",
							Value:            "140101",
						},
						AddressTypeCode: UBLIDWithScheme{
							SchemeAgencyName: "PE:SUNAT",
							SchemeName:       "Establecimientos anexos",
							Value:          "0000",
						},
						CityName:         party.Address.City,
						CountrySubentity: party.Address.Province,
						District:         party.Address.District,
						AddressLine: UBLAddressLine{
							Line: fmt.Sprintf("%s - %s - %s - %s", party.Address.Street, party.Address.District, party.Address.Province, party.Address.Department),
						},
						Country: UBLCountry{
							IdentificationCode: UBLIDWithScheme{
								SchemeAgencyName: "United Nations Economic Commission for Europe",
								SchemeID:         "ISO 3166-1",
								SchemeName:       "Country",
								Value:          party.Address.Country,
							},
						},
					},
				},
			},
			Contact: &UBLContact{
				Name: "",
			},
		},
	}
}

func (c *UBLConverter) convertTaxTotals(taxes []TaxTotal, currency string) []UBLTaxTotal {
	var taxTotals []UBLTaxTotal
	for _, tax := range taxes {
		taxTotal := UBLTaxTotal{
			TaxAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      tax.TaxAmount,
			},
			TaxSubtotals: []UBLTaxSubtotal{
				{
					TaxableAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      tax.TaxBase,
					},
					TaxAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      tax.TaxAmount,
					},
					TaxCategory: UBLTaxCategory{
						ID: UBLIDWithScheme{
							SchemeAgencyName: "United Nations Economic Commission for Europe",
							SchemeID:         "UN/ECE 5305",
							SchemeName:       "Tax Category Identifier",
							Value:            "S",
						},
						Percent: tax.TaxRate,
						TaxExemptionReasonCode: UBLIDWithScheme{
							SchemeAgencyName: "PE:SUNAT",
							SchemeName:       "Afectacion del IGV",
							SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo07",
							Value:          "10",
						},
						TaxScheme: UBLTaxScheme{
							ID: UBLIDWithScheme{
								SchemeAgencyName: "PE:SUNAT",
								SchemeID:         "UN/ECE 5153",
								Value:            tax.TaxType,
							},
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
			Value:      totals.SubTotal,
		},
		TaxInclusiveAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      totals.TotalAmount,
		},
		PayableAmount: UBLAmountWithCurrency{
			CurrencyID: currency,
			Value:      totals.PayableAmount,
		},
	}
}

func (c *UBLConverter) convertInvoiceLines(items []DocumentItem, currency string) []UBLInvoiceLine {
	var lines []UBLInvoiceLine
	for i, item := range items {
		line := UBLInvoiceLine{
			ID: fmt.Sprintf("%d", i+1),
			InvoicedQuantity: UBLQuantityWithUnit{
				UnitCode:                    item.UnitCode,
				UnitCodeListAgencyName:      "United Nations Economic Commission for Europe",
				UnitCodeListID:              "UN/ECE rec 20",
				Value:                       item.Quantity,
			},
			LineExtensionAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      item.LineTotal,
			},
			PricingReference: &UBLPricingReference{
				AlternativeConditionPrice: UBLAlternativeConditionPrice{
					PriceAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      item.UnitPrice * item.Quantity,
					},
					PriceTypeCode: UBLIDWithScheme{
						SchemeAgencyName: "PE:SUNAT",
						SchemeName:       "Tipo de Precio",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo16",
						Value:          "01",
					},
				},
			},
			TaxTotal: c.convertItemTaxes(item.Taxes, currency),
			Item: UBLItem{
				Description: item.Description,
				SellersItemIdentification: &UBLSellersItemIdentification{
					ID: item.ID,
				},
				CommodityClassification: &UBLCommodityClassification{
					ItemClassificationCode: UBLIDWithScheme{
						SchemeAgencyName: "GS1 US",
						SchemeID:         "UNSPSC",
						SchemeName:       "Item Classification",
						Value:          "10191509",
					},
				},
			},
			Price: UBLPrice{
				PriceAmount: UBLAmountWithCurrency{
					CurrencyID: currency,
					Value:      item.UnitPrice,
				},
			},
		}
		lines = append(lines, line)
	}
	return lines
}

func (c *UBLConverter) convertCreditNoteLines(items []DocumentItem, currency string) []UBLCreditNoteLine {
	var lines []UBLCreditNoteLine
	for i, item := range items {
		line := UBLCreditNoteLine{
			ID: fmt.Sprintf("%d", i+1),
			CreditedQuantity: UBLQuantityWithUnit{
				UnitCode:                    item.UnitCode,
				UnitCodeListAgencyName:      "United Nations Economic Commission for Europe",
				UnitCodeListID:              "UN/ECE rec 20",
				Value:                       item.Quantity,
			},
			LineExtensionAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      item.LineTotal,
			},
			PricingReference: &UBLPricingReference{
				AlternativeConditionPrice: UBLAlternativeConditionPrice{
					PriceAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      item.UnitPrice * item.Quantity,
					},
					PriceTypeCode: UBLIDWithScheme{
						SchemeAgencyName: "PE:SUNAT",
						SchemeName:       "Tipo de Precio",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo16",
						Value:          "01",
					},
				},
			},
			TaxTotal: c.convertItemTaxes(item.Taxes, currency),
			Item: UBLItem{
				Description: item.Description,
				SellersItemIdentification: &UBLSellersItemIdentification{
					ID: item.ID,
				},
				CommodityClassification: &UBLCommodityClassification{
					ItemClassificationCode: UBLIDWithScheme{
						SchemeAgencyName: "GS1 US",
						SchemeID:         "UNSPSC",
						SchemeName:       "Item Classification",
						Value:          "10191509",
					},
				},
			},
			Price: UBLPrice{
				PriceAmount: UBLAmountWithCurrency{
					CurrencyID: currency,
					Value:      item.UnitPrice,
				},
			},
		}
		lines = append(lines, line)
	}
	return lines
}

func (c *UBLConverter) convertDebitNoteLines(items []DocumentItem, currency string) []UBLDebitNoteLine {
	var lines []UBLDebitNoteLine
	for i, item := range items {
		line := UBLDebitNoteLine{
			ID: fmt.Sprintf("%d", i+1),
			DebitedQuantity: UBLQuantityWithUnit{
				UnitCode:                    item.UnitCode,
				UnitCodeListAgencyName:      "United Nations Economic Commission for Europe",
				UnitCodeListID:              "UN/ECE rec 20",
				Value:                       item.Quantity,
			},
			LineExtensionAmount: UBLAmountWithCurrency{
				CurrencyID: currency,
				Value:      item.LineTotal,
			},
			PricingReference: &UBLPricingReference{
				AlternativeConditionPrice: UBLAlternativeConditionPrice{
					PriceAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      item.UnitPrice * item.Quantity,
					},
					PriceTypeCode: UBLIDWithScheme{
						SchemeAgencyName: "PE:SUNAT",
						SchemeName:       "Tipo de Precio",
						SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo16",
						Value:          "01",
					},
				},
			},
			TaxTotal: c.convertItemTaxes(item.Taxes, currency),
			Item: UBLItem{
				Description: item.Description,
				SellersItemIdentification: &UBLSellersItemIdentification{
					ID: item.ID,
				},
				CommodityClassification: &UBLCommodityClassification{
					ItemClassificationCode: UBLIDWithScheme{
						SchemeAgencyName: "GS1 US",
						SchemeID:         "UNSPSC",
						SchemeName:       "Item Classification",
						Value:          "10191509",
					},
				},
			},
			Price: UBLPrice{
				PriceAmount: UBLAmountWithCurrency{
					CurrencyID: currency,
					Value:      item.UnitPrice,
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
				Value:      tax.TaxAmount,
			},
			TaxSubtotals: []UBLTaxSubtotal{
				{
					TaxableAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      tax.TaxBase,
					},
					TaxAmount: UBLAmountWithCurrency{
						CurrencyID: currency,
						Value:      tax.TaxAmount,
					},
					TaxCategory: UBLTaxCategory{
						ID: UBLIDWithScheme{
							SchemeAgencyName: "United Nations Economic Commission for Europe",
							SchemeID:         "UN/ECE 5305",
							SchemeName:       "Tax Category Identifier",
							Value:            "S",
						},
						Percent: tax.TaxRate,
						TaxExemptionReasonCode: UBLIDWithScheme{
							SchemeAgencyName: "PE:SUNAT",
							SchemeName:       "Afectacion del IGV",
							SchemeURI:        "urn:pe:gob:sunat:cpe:see:gem:catalogos:catalogo07",
							Value:          "10",
						},
						TaxScheme: UBLTaxScheme{
							ID: UBLIDWithScheme{
								SchemeAgencyName: "PE:SUNAT",
								SchemeID:         "UN/ECE 5153",
								SchemeName:       "Codigo de tributos",
								Value:            tax.TaxType,
							},
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