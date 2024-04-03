package main

import (
	"fmt"
	"os"

	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initializing environment variables
	initializers.InitializeEnv()

	// Starting fiber
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT")))
}