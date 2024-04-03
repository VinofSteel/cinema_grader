package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Calling initializers
	initializers.InitializeEnv()
	db := initializers.InitializeDB()
	defer db.Close()

	// Starting fiber
	fiberConfig := fiber.Config{
		AppName:       "Cinema Grader",
		Prefork:       false,
		CaseSensitive: true,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  90 * time.Second,
		IdleTimeout:   120 * time.Second,
	}
	app := fiber.New(fiberConfig)

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
