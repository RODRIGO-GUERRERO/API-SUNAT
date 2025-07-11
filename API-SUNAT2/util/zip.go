package util

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Empaqueta un archivo XML en un ZIP con el mismo nombre base
func ZipXMLFile(xmlPath string) (string, error) {
	log.Printf("DEBUG: Iniciando empaquetado ZIP para: %s", xmlPath)

	zipPath := xmlPath[:len(xmlPath)-4] + ".zip" // reemplaza .xml por .zip
	log.Printf("DEBUG: Ruta del ZIP será: %s", zipPath)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		log.Printf("ERROR: No se pudo crear el archivo ZIP: %v", err)
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		log.Printf("ERROR: No se pudo abrir el archivo XML: %v", err)
		return "", err
	}
	defer xmlFile.Close()

	xmlName := filepath.Base(xmlPath)
	log.Printf("DEBUG: Nombre del archivo dentro del ZIP: %s", xmlName)

	writer, err := zipWriter.Create(xmlName)
	if err != nil {
		log.Printf("ERROR: No se pudo crear entrada en el ZIP: %v", err)
		return "", err
	}

	_, err = io.Copy(writer, xmlFile)
	if err != nil {
		log.Printf("ERROR: No se pudo copiar contenido al ZIP: %v", err)
		return "", err
	}

	log.Printf("DEBUG: ZIP creado exitosamente en: %s", zipPath)
	return zipPath, nil
} 