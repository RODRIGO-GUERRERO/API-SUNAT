package model

import (
	"encoding/xml"
	"fmt"
	"time"
)

// ============================================================================
// ESTRUCTURAS DE DATOS
// ============================================================================

type BusinessDocument struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Series       string                 `json:"series"`
	Number       string                 `json:"number"`
	IssueDate    string                 `json:"issueDate"`
	DueDate      string                 `json:"dueDate,omitempty"`
	Currency     string                 `json:"currency"`
	Issuer       Party                  `json:"issuer"`
	Customer     Party                  `json:"customer"`
	Items        []DocumentItem         `json:"items"`
	Totals       DocumentTotals         `json:"totals"`
	Taxes        []TaxTotal             `json:"taxes"`
	Additional   map[string]interface{} `json:"additional,omitempty"`
	Reference    *DocumentReference     `json:"reference,omitempty"`
}

type Party struct {
	DocumentType string `json:"documentType"`
	DocumentID   string `json:"documentId"`
	Name         string `json:"name"`
	TradeName    string `json:"tradeName,omitempty"`
	Address      Address `json:"address"`
}

type Address struct {
	Street      string `json:"street"`
	City        string `json:"city"`
	District    string `json:"district"`
	Province    string `json:"province"`
	Department  string `json:"department"`
	Country     string `json:"country"`
	PostalCode  string `json:"postalCode,omitempty"`
}

type DocumentItem struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitCode    string  `json:"unitCode"`
	UnitPrice   float64 `json:"unitPrice"`
	LineTotal   float64 `json:"lineTotal"`
	Taxes       []Tax   `json:"taxes"`
}

type DocumentTotals struct {
	SubTotal     float64 `json:"subTotal"`
	TotalTaxes   float64 `json:"totalTaxes"`
	TotalAmount  float64 `json:"totalAmount"`
	PayableAmount float64 `json:"payableAmount"`
}

type TaxTotal struct {
	TaxType   string  `json:"taxType"`
	TaxAmount float64 `json:"taxAmount"`
	TaxRate   float64 `json:"taxRate,omitempty"`
	TaxBase   float64 `json:"taxBase,omitempty"`
}

type Tax struct {
	TaxType   string  `json:"taxType"`
	TaxAmount float64 `json:"taxAmount"`
	TaxRate   float64 `json:"taxRate,omitempty"`
	TaxBase   float64 `json:"taxBase,omitempty"`
}

type DocumentReference struct {
	DocumentType string `json:"documentType"`
	DocumentID   string `json:"documentId"`
	IssueDate    string `json:"issueDate"`
	Reason       string `json:"reason"`
}

// Estructuras UBL 2.1 XML
type UBLInvoice struct {
	XMLName                xml.Name               `xml:"Invoice"`
	Xmlns                  string                 `xml:"xmlns,attr"`
	XmlnsCbc               string                 `xml:"xmlns:cbc,attr"`
	XmlnsCac               string                 `xml:"xmlns:cac,attr"`
	XmlnsExt               string                 `xml:"xmlns:ext,attr"`
	XmlnsDs                string                 `xml:"xmlns:ds,attr"`
	UBLExtensions          *UBLExtensions         `xml:"ext:UBLExtensions"`
	UBLVersionID           string                 `xml:"cbc:UBLVersionID"`
	CustomizationID        UBLIDWithScheme        `xml:"cbc:CustomizationID"`
	ProfileID              UBLIDWithScheme        `xml:"cbc:ProfileID"`
	ID                     string                 `xml:"cbc:ID"`
	IssueDate              string                 `xml:"cbc:IssueDate"`
	IssueTime              string                 `xml:"cbc:IssueTime,omitempty"`
	DueDate                string                 `xml:"cbc:DueDate,omitempty"`
	InvoiceTypeCode        UBLTypeCode            `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode   UBLIDWithScheme        `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric       int                    `xml:"cbc:LineCountNumeric"`
	AccountingSupplierParty UBLParty              `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty UBLParty              `xml:"cac:AccountingCustomerParty"`
	PaymentTerms           []UBLPaymentTerms      `xml:"cac:PaymentTerms,omitempty"`
	TaxTotal               []UBLTaxTotal          `xml:"cac:TaxTotal"`
	LegalMonetaryTotal     UBLLegalMonetaryTotal  `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines           []UBLInvoiceLine       `xml:"cac:InvoiceLine"`
}

type UBLCreditNote struct {
	XMLName                xml.Name               `xml:"CreditNote"`
	Xmlns                  string                 `xml:"xmlns,attr"`
	XmlnsCac               string                 `xml:"xmlns:cac,attr"`
	XmlnsCbc               string                 `xml:"xmlns:cbc,attr"`
	XmlnsDs                string                 `xml:"xmlns:ds,attr"`
	XmlnsExt               string                 `xml:"xmlns:ext,attr"`
	UBLExtensions          *UBLExtensions         `xml:"ext:UBLExtensions,omitempty"`
	UBLVersionID           string                 `xml:"cbc:UBLVersionID"`
	CustomizationID        UBLIDWithScheme        `xml:"cbc:CustomizationID"`
	ProfileID              UBLIDWithScheme        `xml:"cbc:ProfileID"`
	ID                     string                 `xml:"cbc:ID"`
	IssueDate              string                 `xml:"cbc:IssueDate"`
	IssueTime              string                 `xml:"cbc:IssueTime,omitempty"`
	CreditNoteTypeCode     UBLTypeCode            `xml:"cbc:CreditNoteTypeCode"`
	DocumentCurrencyCode   UBLIDWithScheme        `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric       int                    `xml:"cbc:LineCountNumeric"`
	DiscrepancyResponse    []UBLDiscrepancyResponse `xml:"cac:DiscrepancyResponse"`
	BillingReference       []UBLBillingReference  `xml:"cac:BillingReference"`
	AccountingSupplierParty UBLParty              `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty UBLParty              `xml:"cac:AccountingCustomerParty"`
	PaymentTerms           []UBLPaymentTerms      `xml:"cac:PaymentTerms,omitempty"`
	TaxTotal               []UBLTaxTotal          `xml:"cac:TaxTotal"`
	LegalMonetaryTotal     UBLLegalMonetaryTotal  `xml:"cac:LegalMonetaryTotal"`
	CreditNoteLines        []UBLCreditNoteLine    `xml:"cac:CreditNoteLine"`
}

type UBLDebitNote struct {
	XMLName                xml.Name               `xml:"DebitNote"`
	Xmlns                  string                 `xml:"xmlns,attr"`
	XmlnsCac               string                 `xml:"xmlns:cac,attr"`
	XmlnsCbc               string                 `xml:"xmlns:cbc,attr"`
	XmlnsDs                string                 `xml:"xmlns:ds,attr"`
	XmlnsExt               string                 `xml:"xmlns:ext,attr"`
	UBLExtensions          *UBLExtensions         `xml:"ext:UBLExtensions,omitempty"`
	UBLVersionID           string                 `xml:"cbc:UBLVersionID"`
	CustomizationID        UBLIDWithScheme        `xml:"cbc:CustomizationID"`
	ProfileID              UBLIDWithScheme        `xml:"cbc:ProfileID"`
	ID                     string                 `xml:"cbc:ID"`
	IssueDate              string                 `xml:"cbc:IssueDate"`
	IssueTime              string                 `xml:"cbc:IssueTime,omitempty"`
	DebitNoteTypeCode      UBLTypeCode            `xml:"cbc:DebitNoteTypeCode"`
	DocumentCurrencyCode   UBLIDWithScheme        `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric       int                    `xml:"cbc:LineCountNumeric"`
	DiscrepancyResponse    []UBLDiscrepancyResponse `xml:"cac:DiscrepancyResponse"`
	BillingReference       []UBLBillingReference  `xml:"cac:BillingReference"`
	AccountingSupplierParty UBLParty              `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty UBLParty              `xml:"cac:AccountingCustomerParty"`
	PaymentTerms           []UBLPaymentTerms      `xml:"cac:PaymentTerms,omitempty"`
	TaxTotal               []UBLTaxTotal          `xml:"cac:TaxTotal"`
	LegalMonetaryTotal     UBLLegalMonetaryTotal  `xml:"cac:LegalMonetaryTotal"`
	DebitNoteLines         []UBLDebitNoteLine     `xml:"cac:DebitNoteLine"`
}

type UBLExtensions struct {
	UBLExtension UBLExtension `xml:"ext:UBLExtension"`
}

type UBLExtension struct {
	ExtensionContent ExtensionContent `xml:"ext:ExtensionContent"`
}

type ExtensionContent struct {
	Signature XMLSignature `xml:"ds:Signature"`
}

// Modificar UBLParty para SUNAT
type UBLParty struct {
	CustomerAssignedAccountID string `xml:"cbc:CustomerAssignedAccountID"`
	AdditionalAccountID       string `xml:"cbc:AdditionalAccountID"`
	PartyIdentification      UBLPartyIdentification `xml:"cac:PartyIdentification"`
	Party                    UBLPartyDetail `xml:"cac:Party"`
}

type UBLPartyDetail struct {
	PartyName           []UBLPartyName           `xml:"cac:PartyName"`
	RegistrationAddress UBLRegistrationAddress   `xml:"cac:RegistrationAddress"`
	PartyTaxScheme      []UBLPartyTaxScheme      `xml:"cac:PartyTaxScheme"`
	PartyLegalEntity    []UBLPartyLegalEntity    `xml:"cac:PartyLegalEntity"`
	Contact             *UBLContact              `xml:"cac:Contact,omitempty"`
}

type UBLPartyIdentification struct {
	ID UBLIDWithScheme `xml:"cbc:ID"`
}

type UBLIDWithScheme struct {
	SchemeAgencyName string `xml:"schemeAgencyName,attr,omitempty"`
	SchemeID         string `xml:"schemeID,attr,omitempty"`
	SchemeName       string `xml:"schemeName,attr,omitempty"`
	SchemeURI        string `xml:"schemeURI,attr,omitempty"`
	Value            string `xml:",chardata"`
}

type UBLPartyName struct {
	Name string `xml:"cbc:Name"`
}

type UBLRegistrationAddress struct {
	ID                 UBLIDWithScheme `xml:"cbc:ID,omitempty"`
	AddressTypeCode    UBLIDWithScheme `xml:"cbc:AddressTypeCode,omitempty"`
	CityName           string          `xml:"cbc:CityName"`
	CountrySubentity   string          `xml:"cbc:CountrySubentity"`
	District           string          `xml:"cbc:District"`
	AddressLine        UBLAddressLine  `xml:"cac:AddressLine"`
	Country            UBLCountry      `xml:"cac:Country"`
}

type UBLAddressLine struct {
	Line string `xml:"cbc:Line"`
}

type UBLCountry struct {
	IdentificationCode UBLIDWithScheme `xml:"cbc:IdentificationCode"`
}

type UBLPartyTaxScheme struct {
	RegistrationName string          `xml:"cbc:RegistrationName"`
	CompanyID        UBLIDWithScheme `xml:"cbc:CompanyID"`
	TaxScheme        UBLTaxScheme    `xml:"cac:TaxScheme"`
}

type UBLTaxScheme struct {
	ID           UBLIDWithScheme `xml:"cbc:ID"`
	Name         string          `xml:"cbc:Name,omitempty"`
	TaxTypeCode  string          `xml:"cbc:TaxTypeCode,omitempty"`
}

type UBLPartyLegalEntity struct {
	RegistrationName     string                `xml:"cbc:RegistrationName"`
	RegistrationAddress  UBLRegistrationAddress `xml:"cac:RegistrationAddress"`
}

type UBLContact struct {
	Name string `xml:"cbc:Name,omitempty"`
}

type UBLTaxTotal struct {
	TaxAmount     UBLAmountWithCurrency `xml:"cbc:TaxAmount"`
	TaxSubtotals  []UBLTaxSubtotal      `xml:"cac:TaxSubtotal"`
}

// Tipo para serializar siempre con dos decimales
type Decimal2 float64

func (d Decimal2) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	value := fmt.Sprintf("%.2f", float64(d))
	return e.EncodeElement(value, start)
}

type UBLAmountWithCurrency struct {
	CurrencyID string   `xml:"currencyID,attr"`
	Value      Decimal2 `xml:",chardata"`
}

type UBLTaxSubtotal struct {
	TaxableAmount UBLAmountWithCurrency `xml:"cbc:TaxableAmount"`
	TaxAmount     UBLAmountWithCurrency `xml:"cbc:TaxAmount"`
	TaxCategory   UBLTaxCategory        `xml:"cac:TaxCategory"`
}

type UBLTaxCategory struct {
	ID                    UBLIDWithScheme `xml:"cbc:ID"`
	Percent               float64         `xml:"cbc:Percent,omitempty"`
	TaxExemptionReasonCode UBLIDWithScheme `xml:"cbc:TaxExemptionReasonCode,omitempty"`
	TaxScheme             UBLTaxScheme    `xml:"cac:TaxScheme"`
}

type UBLLegalMonetaryTotal struct {
	LineExtensionAmount UBLAmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	TaxInclusiveAmount  UBLAmountWithCurrency `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       UBLAmountWithCurrency `xml:"cbc:PayableAmount"`
}

type UBLInvoiceLine struct {
	ID                  string                `xml:"cbc:ID"`
	InvoicedQuantity    UBLQuantityWithUnit   `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount UBLAmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	PricingReference    *UBLPricingReference  `xml:"cac:PricingReference,omitempty"`
	TaxTotal            []UBLTaxTotal         `xml:"cac:TaxTotal"`
	Item                UBLItem               `xml:"cac:Item"`
	Price               UBLPrice              `xml:"cac:Price"`
}

type UBLCreditNoteLine struct {
	ID                    string                `xml:"cbc:ID"`
	CreditedQuantity      UBLQuantityWithUnit   `xml:"cbc:CreditedQuantity"`
	LineExtensionAmount   UBLAmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	PricingReference      *UBLPricingReference  `xml:"cac:PricingReference,omitempty"`
	TaxTotal              []UBLTaxTotal         `xml:"cac:TaxTotal"`
	Item                  UBLItem               `xml:"cac:Item"`
	Price                 UBLPrice              `xml:"cac:Price"`
}

type UBLQuantityWithUnit struct {
	UnitCode                    string   `xml:"unitCode,attr"`
	UnitCodeListAgencyName      string   `xml:"unitCodeListAgencyName,attr,omitempty"`
	UnitCodeListID              string   `xml:"unitCodeListID,attr,omitempty"`
	Value                       Decimal2 `xml:",chardata"`
}

type UBLPricingReference struct {
	AlternativeConditionPrice UBLAlternativeConditionPrice `xml:"cac:AlternativeConditionPrice"`
}

type UBLAlternativeConditionPrice struct {
	PriceAmount   UBLAmountWithCurrency `xml:"cbc:PriceAmount"`
	PriceTypeCode UBLIDWithScheme       `xml:"cbc:PriceTypeCode"`
}

type UBLPrice struct {
	PriceAmount UBLAmountWithCurrency `xml:"cbc:PriceAmount"`
}

type UBLItem struct {
	Description              string                    `xml:"cbc:Description"`
	SellersItemIdentification *UBLSellersItemIdentification `xml:"cac:SellersItemIdentification,omitempty"`
	CommodityClassification  *UBLCommodityClassification  `xml:"cac:CommodityClassification,omitempty"`
}

type UBLSellersItemIdentification struct {
	ID string `xml:"cbc:ID"`
}

type UBLCommodityClassification struct {
	ItemClassificationCode UBLIDWithScheme `xml:"cbc:ItemClassificationCode"`
}

type UBLDiscrepancyResponse struct {
	ReferenceID string `xml:"cbc:ReferenceID"`
	ResponseCode string `xml:"cbc:ResponseCode"`
	Description string `xml:"cbc:Description"`
}

type UBLBillingReference struct {
	InvoiceDocumentReference UBLDocumentReference `xml:"cac:InvoiceDocumentReference"`
}

type UBLDocumentReference struct {
	ID        string `xml:"cbc:ID"`
	IssueDate string `xml:"cbc:IssueDate"`
	DocumentTypeCode string `xml:"cbc:DocumentTypeCode"`
}

type UBLTypeCode struct {
	ListAgencyName string `xml:"listAgencyName,attr,omitempty"`
	ListID         string `xml:"listID,attr,omitempty"`
	ListName       string `xml:"listName,attr,omitempty"`
	ListURI        string `xml:"listURI,attr,omitempty"`
	Name           string `xml:"name,attr,omitempty"`
	Value          string `xml:",chardata"`
}

type UBLPaymentTerms struct {
	ID              string `xml:"cbc:ID"`
	PaymentMeansID  string `xml:"cbc:PaymentMeansID"`
}

type UBLDelivery struct {
	DeliveryDate string `xml:"cbc:DeliveryDate"`
}

type UBLDebitNoteLine struct {
	ID                    string                `xml:"cbc:ID"`
	DebitedQuantity       UBLQuantityWithUnit   `xml:"cbc:DebitedQuantity"`
	LineExtensionAmount   UBLAmountWithCurrency `xml:"cbc:LineExtensionAmount"`
	PricingReference      *UBLPricingReference  `xml:"cac:PricingReference,omitempty"`
	TaxTotal              []UBLTaxTotal         `xml:"cac:TaxTotal"`
	Item                  UBLItem               `xml:"cac:Item"`
	Price                 UBLPrice              `xml:"cac:Price"`
}

// Estructura para la firma digital
type XMLSignature struct {
	XMLName         xml.Name        `xml:"ds:Signature"`
	Id              string          `xml:"Id,attr"`
	SignedInfo      SignedInfo      `xml:"ds:SignedInfo"`
	SignatureValue  SignatureValue  `xml:"ds:SignatureValue"`
	KeyInfo         KeyInfo         `xml:"ds:KeyInfo"`
}

type SignedInfo struct {
	CanonicalizationMethod CanonicalizationMethod `xml:"ds:CanonicalizationMethod"`
	SignatureMethod        SignatureMethod        `xml:"ds:SignatureMethod"`
	Reference              Reference              `xml:"ds:Reference"`
}

type CanonicalizationMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type SignatureMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type Reference struct {
	URI          string       `xml:"URI,attr"`
	Transforms   Transforms   `xml:"ds:Transforms"`
	DigestMethod DigestMethod `xml:"ds:DigestMethod"`
	DigestValue  string       `xml:"ds:DigestValue"`
}

type Transforms struct {
	Transform []Transform `xml:"ds:Transform"`
}

type Transform struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type DigestMethod struct {
	Algorithm string `xml:"Algorithm,attr"`
}

type SignatureValue struct {
	Value string `xml:",chardata"`
}

type KeyInfo struct {
	X509Data X509Data `xml:"ds:X509Data"`
}

type X509Data struct {
	X509Certificate string `xml:"ds:X509Certificate"`
}

// Eliminar la definición de UBLSignature y sus dependencias

type APIResponse struct {
	Status        string                 `json:"status"`
	CorrelationID string                 `json:"correlationId"`
	DocumentID    string                 `json:"documentId,omitempty"`
	XMLPath       string                 `json:"xmlPath,omitempty"`
	XMLHash       string                 `json:"xmlHash,omitempty"`
	ProcessedAt   time.Time              `json:"processedAt"`
	Duration      int64                  `json:"duration,omitempty"`
	ErrorCode     string                 `json:"errorCode,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	ValidationErrors []ValidationError   `json:"validationErrors,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Message       string                 `json:"message,omitempty"`
}

type ValidationError struct {
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Received string `json:"received"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}