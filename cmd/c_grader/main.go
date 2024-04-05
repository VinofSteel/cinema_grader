package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/gofiber/fiber/v2"
)

type GlobalErrorHandlerResp struct {
	Message string `json:"message"`
}

func main() {
	// Calling initializers
	initializers.InitializeEnv()
	initializers.InitializeValidator()
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
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Message: err.Error(),
			})
		},
	}
	app := fiber.New(fiberConfig)

	// Controllers
	userController := controllers.User{
		DB: db,
	}

	// Routes - User
	app.Post("/users", userController.CreateUser)

	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
