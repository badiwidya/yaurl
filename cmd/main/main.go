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
	address := cfg.APP_HOST + ":" + cfg.APP_PORT

	app := app.New(address, cfg)

	if err := app.Run(); err != nil {
		log.Fatalf("Failed to start server")
	}
}
