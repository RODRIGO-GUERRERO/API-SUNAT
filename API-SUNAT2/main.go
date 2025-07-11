package main

import (
	"log"

	"API-SUNAT2/api"
	"API-SUNAT2/config"
)

func main() {
	cfg := config.LoadConfig()
	router := api.NewRouter()

	log.Printf("Servidor iniciado en el puerto %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
} 