package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

func InitializeEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error initializing environment variables")
	}
}