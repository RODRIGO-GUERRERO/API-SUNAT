package service

import (
	"crypto"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	. "API-SUNAT2/model"
	"github.com/sirupsen/logrus"
)

type DigitalSignatureService struct {
	logger *logrus.Logger
}

func NewDigitalSignatureService(logger *logrus.Logger) *DigitalSignatureService {
	return &DigitalSignatureService{logger: logger}
}

func (s *DigitalSignatureService) SignXML(xmlContent []byte, certPEM []byte, keyPEM []byte) ([]byte, error) {
	// Decodificar certificado y clave privada
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode private key PEM")
	}

	// Intentar parsear como PKCS1 primero
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		// Si falla, intentar como PKCS8
		key, err2 := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse private key (PKCS1: %v, PKCS8: %v)", err, err2)
		}

		// Convertir a *rsa.PrivateKey
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		privateKey = rsaKey
	}

	// Generar hash SHA-256 del contenido XML
	hash := sha256.Sum256(xmlContent)

	// Firmar el hash
	signature, err := rsa.SignPKCS1v15(cryptorand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %v", err)
	}

	// Codificar firma en base64
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	// Codificar certificado en base64
	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	// Crear estructura XMLDSig
	xmlSignature := &XMLSignature{
		Id: "SignatureSP",
		SignedInfo: SignedInfo{
			CanonicalizationMethod: CanonicalizationMethod{
				Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
			},
			SignatureMethod: SignatureMethod{
				Algorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
			},
			Reference: Reference{
				URI: "",
				Transforms: Transforms{
					Transform: []Transform{
						{Algorithm: "http://www.w3.org/2000/09/xmldsig#enveloped-signature"},
						{Algorithm: "http://www.w3.org/2001/10/xml-exc-c14n#"},
					},
				},
				DigestMethod: DigestMethod{
					Algorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
				},
				DigestValue: base64.StdEncoding.EncodeToString(hash[:]),
			},
		},
		SignatureValue: SignatureValue{
			Value: signatureBase64,
		},
		KeyInfo: KeyInfo{
			X509Data: X509Data{
				X509Certificate: certBase64,
			},
		},
	}

	// Insertar la firma en el XML
	signedXML, err := s.insertSignatureInXML(xmlContent, xmlSignature)
	if err != nil {
		return nil, fmt.Errorf("failed to insert signature: %v", err)
	}

	return signedXML, nil
}

func (s *DigitalSignatureService) insertSignatureInXML(xmlContent []byte, xmlSignature *XMLSignature) ([]byte, error) {
	// Convertir XML a string para manipulación
	xmlStr := string(xmlContent)

	// Crear el bloque UBLExtensions completo con la firma XMLDSig
	extensionsXML := fmt.Sprintf(`<UBLExtensions xmlns="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2">
  <UBLExtension>
    <ExtensionContent>
      <Signature xmlns="http://www.w3.org/2000/09/xmldsig#" Id="%s">
        <SignedInfo>
          <CanonicalizationMethod Algorithm="%s"/>
          <SignatureMethod Algorithm="%s"/>
          <Reference URI="%s">
            <Transforms>
              <Transform Algorithm="%s"/>
              <Transform Algorithm="%s"/>
            </Transforms>
            <DigestMethod Algorithm="%s"/>
            <DigestValue>%s</DigestValue>
          </Reference>
        </SignedInfo>
        <SignatureValue>%s</SignatureValue>
        <KeyInfo>
          <X509Data>
            <X509Certificate>%s</X509Certificate>
          </X509Data>
        </KeyInfo>
      </Signature>
    </ExtensionContent>
  </UBLExtension>
</UBLExtensions>`,
		xmlSignature.Id,
		xmlSignature.SignedInfo.CanonicalizationMethod.Algorithm,
		xmlSignature.SignedInfo.SignatureMethod.Algorithm,
		xmlSignature.SignedInfo.Reference.URI,
		xmlSignature.SignedInfo.Reference.Transforms.Transform[0].Algorithm,
		xmlSignature.SignedInfo.Reference.Transforms.Transform[1].Algorithm,
		xmlSignature.SignedInfo.Reference.DigestMethod.Algorithm,
		xmlSignature.SignedInfo.Reference.DigestValue,
		xmlSignature.SignatureValue.Value,
		xmlSignature.KeyInfo.X509Data.X509Certificate,
	)

	// Buscar el bloque UBLExtensions existente (con namespace completo)
	startTag := "<UBLExtensions"
	endTag := "</UBLExtensions>"
	startIdx := strings.Index(xmlStr, startTag)
	endIdx := strings.Index(xmlStr, endTag)
	
	if startIdx == -1 || endIdx == -1 {
		// Si no encuentra con namespace completo, intentar con prefijo ext:
		startTag = "<ext:UBLExtensions"
		endTag = "</ext:UBLExtensions>"
		startIdx = strings.Index(xmlStr, startTag)
		endIdx = strings.Index(xmlStr, endTag)
		
		if startIdx == -1 || endIdx == -1 {
			return nil, fmt.Errorf("No se encontró el bloque UBLExtensions en el XML")
		}
	}
	
	// Encontrar el final del tag de apertura
	tagEndIdx := strings.Index(xmlStr[startIdx:], ">")
	if tagEndIdx == -1 {
		return nil, fmt.Errorf("No se encontró el cierre del tag UBLExtensions")
	}
	startIdx += tagEndIdx + 1
	
	endIdx += len(endTag)
	
	// Reemplazar el bloque existente por el firmado
	replacedXML := xmlStr[:startIdx] + extensionsXML + xmlStr[endIdx:]
	return []byte(replacedXML), nil
} 