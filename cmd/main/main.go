package main

import (
	"log"

	"github.com/badiwidya/yaurl/internal/app"
	"github.com/badiwidya/yaurl/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load dotenv: %v\n", err)
	}
	cfg := config.New()

	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v\n", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("Server stopped with error: %v\n", err)
	}
}
