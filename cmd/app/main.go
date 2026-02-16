package main

import (
	"log"
	"os"

	"github.com/andreyxaxa/Image-Processor/config"
	"github.com/andreyxaxa/Image-Processor/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	// Config
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			log.Fatalf("config error: %s", err)
		}
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
