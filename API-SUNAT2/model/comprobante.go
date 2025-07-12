package model

import (
	"encoding/xml"
	"fmt"
	"time"
)

// ============================================================================
// ESTRUCTURAS DE DATOS JSON (SIN CAMBIOS)
// ============================================================================

type BusinessDocument struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Series    string         `json:"series"`
	Number    string         `json:"number"`
	IssueDate string         `json:"issueDate"`
	DueDate   string         `json:"dueDate,omitempty"`
	Currency  string         `json:"currency"`
	Issuer    Party          `json:"issuer"`
	Customer  Party          `json:"customer"`
	Items     []DocumentItem `json:"items"`
	Totals    DocumentTotals `json:"totals"`
	Taxes     []TaxTotal     `json:"taxes"`
	Additional map[string]interface{} `json:"additional,omitempty"`
	Reference *DocumentReference     `json:"reference,omitempty"`
}

type Party struct {
	DocumentType string  `json:"documentType"`
	DocumentID   string  `json:"documentId"`
	Name         string  `json:"name"`
	TradeName    string  `json:"tradeName,omitempty"`
	Address      Address `json:"address"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	District   string `json:"district"`
	Province   string `json:"province"`
	Department string `json:"department"`
	Country    string `json:"country"`
	PostalCode string `json:"postalCode,omitempty"`
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
	SubTotal      float64 `json:"subTotal"`
	TotalTaxes    float64 `json:"totalTaxes"`
	TotalAmount   float64 `json:"totalAmount"`
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

// ============================================================================
// ESTRUCTURAS UBL 2.1 XML (CORREGIDAS)
// ============================================================================

// ---
// Estructura Raíz de la Factura (Invoice)
// Se han añadido los atributos xmlns para definir los prefijos una sola vez.
// ---
type UBLInvoice struct {
	XMLName                xml.Name `xml:"Invoice"`
	Xmlns                  string   `xml:"xmlns,attr"`
	XmlnsCac               string   `xml:"xmlns:cac,attr"`
	XmlnsCbc               string   `xml:"xmlns:cbc,attr"`
	XmlnsDs                string   `xml:"xmlns:ds,attr"`
	XmlnsExt               string   `xml:"xmlns:ext,attr"`

	UBLExtensions          *UBLExtensions        `xml:"ext:UBLExtensions"`
	UBLVersionID           string                `xml:"cbc:UBLVersionID"`
	CustomizationID        string                `xml:"cbc:CustomizationID"`
	ProfileID              string                `xml:"cbc:ProfileID"`
	ID                     string                `xml:"cbc:ID"`
	IssueDate              string                `xml:"cbc:IssueDate"`
	IssueTime              string                `xml:"cbc:IssueTime,omitempty"`
	DueDate                string                `xml:"cbc:DueDate,omitempty"`
	InvoiceTypeCode        UBLTypeCode           `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode   UBLCurrencyCode       `xml:"cbc:DocumentCurrencyCode"`
	LineCountNumeric       int                   `xml:"cbc:LineCountNumeric"`
	Signature              *UBLSignature         `xml:"cac:Signature"`
	AccountingSupplierParty *UBLParty             `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty *UBLParty             `xml:"cac:AccountingCustomerParty"`
	PaymentTerms           []UBLPaymentTerms     `xml:"cac:PaymentTerms,omitempty"`
	TaxTotal               []UBLTaxTotal         `xml:"cac:TaxTotal"`
	LegalMonetaryTotal     UBLLegalMonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLines           []UBLInvoiceLine      `xml:"cac:InvoiceLine"`
}


// NOTA: No he corregido UBLCreditNote y UBLDebitNote, pero deberás aplicar
// la misma lógica que en UBLInvoice si las vas a usar.
type UBLCreditNote struct {
    // ... aplicar la misma corrección de namespaces que en UBLInvoice ...
}
type UBLDebitNote struct {
    // ... aplicar la misma corrección de namespaces que en UBLInvoice ...
}


// ---
// Secciones de la Factura
// Todas las etiquetas ahora usan prefijos (cbc:, cac:, etc.) en lugar de la URL completa.
// ---

type UBLExtensions struct {
	XMLName       xml.Name       `xml:"ext:UBLExtensions"`
	UBLExtension  []UBLExtension `xml:"ext:UBLExtension"`
}

type UBLExtension struct {
	XMLName          xml.Name         `xml:"ext:UBLExtension"`
	ExtensionContent ExtensionContent `xml:"ext:ExtensionContent"`
}

type ExtensionContent struct {
	XMLName   xml.Name      `xml:"ext:ExtensionContent"`
	Signature *XMLSignature `xml:"ds:Signature"`
}

type UBLParty struct {
	Party UBLPartyDetail `xml:"cac:Party"`
}

type UBLPartyDetail struct {
	PartyIdentification UBLPartyIdentification `xml:"cac:PartyIdentification"`
	PartyName           []UBLPartyName         `xml:"cac:PartyName"`
	PartyLegalEntity    []UBLPartyLegalEntity  `xml:"cac:PartyLegalEntity"`
	Contact             *UBLContact            `xml:"cac:Contact,omitempty"`
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
	ID               UBLIDWithScheme `xml:"cbc:ID,omitempty"`
	AddressTypeCode  UBLIDWithScheme `xml:"cbc:AddressTypeCode,omitempty"`
	CityName         string          `xml:"cbc:CityName"`
	CountrySubentity string          `xml:"cbc:CountrySubentity"`
	District         string          `xml:"cbc:District"`
	AddressLine      UBLAddressLine  `xml:"cac:AddressLine"`
	Country          UBLCountry      `xml:"cac:Country"`
}

type UBLAddressLine struct {
	Line string `xml:"cbc:Line"`
}

type UBLCountry struct {
	IdentificationCode UBLIDWithScheme `xml:"cbc:IdentificationCode"`
}

type UBLPartyTaxScheme struct {
	RegistrationName string       `xml:"cbc:RegistrationName"`
	CompanyID        UBLIDWithScheme `xml:"cbc:CompanyID"`
	TaxScheme        UBLTaxScheme `xml:"cac:TaxScheme"`
}

type UBLTaxScheme struct {
	ID          UBLIDWithScheme `xml:"cbc:ID"`
	Name        string          `xml:"cbc:Name,omitempty"`
	TaxTypeCode string          `xml:"cbc:TaxTypeCode,omitempty"`
}

type UBLPartyLegalEntity struct {
	RegistrationName    string               `xml:"cbc:RegistrationName"`
	RegistrationAddress UBLRegistrationAddress `xml:"cac:RegistrationAddress"`
}

type UBLContact struct {
	Name string `xml:"cbc:Name,omitempty"`
}

type UBLTaxTotal struct {
	TaxAmount    UBLAmountWithCurrency `xml:"cbc:TaxAmount"`
	TaxSubtotals []UBLTaxSubtotal      `xml:"cac:TaxSubtotal"`
}

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
	ID                     UBLIDWithScheme `xml:"cbc:ID"`
	Percent                float64         `xml:"cbc:Percent,omitempty"`
	TaxExemptionReasonCode UBLIDWithScheme `xml:"cbc:TaxExemptionReasonCode,omitempty"`
	TaxScheme              UBLTaxScheme    `xml:"cac:TaxScheme"`
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

type UBLQuantityWithUnit struct {
	UnitCode               string   `xml:"unitCode,attr"`
	UnitCodeListAgencyName string   `xml:"unitCodeListAgencyName,attr,omitempty"`
	UnitCodeListID         string   `xml:"unitCodeListID,attr,omitempty"`
	Value                  Decimal2 `xml:",chardata"`
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
	Description             string                       `xml:"cbc:Description"`
	SellersItemIdentification *UBLSellersItemIdentification `xml:"cac:SellersItemIdentification,omitempty"`
	CommodityClassification *UBLCommodityClassification  `xml:"cac:CommodityClassification,omitempty"`
}

type UBLSellersItemIdentification struct {
	ID string `xml:"cbc:ID"`
}

type UBLCommodityClassification struct {
	ItemClassificationCode UBLIDWithScheme `xml:"cbc:ItemClassificationCode"`
}

type UBLTypeCode struct {
	ListAgencyName string `xml:"listAgencyName,attr,omitempty"`
	ListID         string `xml:"listID,attr,omitempty"`
	ListName       string `xml:"listName,attr,omitempty"`
	ListURI        string `xml:"listURI,attr,omitempty"`
	Name           string `xml:"name,attr,omitempty"`
	Value          string `xml:",chardata"`
}

type UBLCurrencyCode struct {
	ListAgencyName string `xml:"listAgencyName,attr,omitempty"`
	ListID         string `xml:"listID,attr,omitempty"`
	ListName       string `xml:"listName,attr,omitempty"`
	Value          string `xml:",chardata"`
}

type UBLPaymentTerms struct {
	ID             string `xml:"cbc:ID"`
	PaymentMeansID string `xml:"cbc:PaymentMeansID"`
}

// Estructuras para la sección <cac:Signature> (distinta de la firma digital <ds:Signature>)
type UBLSignature struct {
	XMLName                    xml.Name `xml:"cac:Signature"`
	ID                         string   `xml:"cbc:ID"`
	SignatoryParty             UBLSignatoryParty `xml:"cac:SignatoryParty"`
	DigitalSignatureAttachment UBLDigitalSignatureAttachment `xml:"cac:DigitalSignatureAttachment"`
}

type UBLSignatoryParty struct {
	PartyIdentification UBLPartyIdentification `xml:"cac:PartyIdentification"`
	PartyName           UBLPartyName           `xml:"cac:PartyName"`
}

type UBLDigitalSignatureAttachment struct {
	ExternalReference UBLExternalReference `xml:"cac:ExternalReference"`
}

type UBLExternalReference struct {
	URI string `xml:"cbc:URI"`
}

// Estructura para la firma digital <ds:Signature>
type XMLSignature struct {
	XMLName        xml.Name       `xml:"ds:Signature"`
	Id             string         `xml:"Id,attr"`
	SignedInfo     SignedInfo     `xml:"ds:SignedInfo"`
	SignatureValue SignatureValue `xml:"ds:SignatureValue"`
	KeyInfo        KeyInfo        `xml:"ds:KeyInfo"`
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


// --- Estructuras para la Respuesta de la API ---

type APIResponse struct {
	Status           string                 `json:"status"`
	CorrelationID    string                 `json:"correlationId"`
	DocumentID       string                 `json:"documentId,omitempty"`
	XMLPath          string                 `json:"xmlPath,omitempty"`
	XMLHash          string                 `json:"xmlHash,omitempty"`
	ProcessedAt      time.Time              `json:"processedAt"`
	Duration         int64                  `json:"duration,omitempty"`
	ErrorCode        string                 `json:"errorCode,omitempty"`
	ErrorMessage     string                 `json:"errorMessage,omitempty"`
	ValidationErrors []ValidationError      `json:"validationErrors,omitempty"`
	Data             map[string]interface{} `json:"data,omitempty"`
	Message          string                 `json:"message,omitempty"`
}

type ValidationError struct {
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Received string `json:"received"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}