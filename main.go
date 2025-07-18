package main

import (
	"os"

	"github.com/cloudbees-days/fm-actions-container/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development and testing)
	// Silently ignore if .env doesn't exist - normal in production
	godotenv.Load()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
