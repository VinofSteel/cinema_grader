package main

import (
	"fmt"
	"os"
	"time"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initializing environment variables
	initializers.InitializeEnv()

	// Starting fiber
	fiberConfig := fiber.Config{
		AppName: "Cinema Grader",
		Prefork: true,
		CaseSensitive: true,
		ReadTimeout: 30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout: 120 * time.Second,
	}
	app := fiber.New(fiberConfig)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT")))
}