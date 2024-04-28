package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/VinOfSteel/cinemagrader/initializers"
	"github.com/VinOfSteel/cinemagrader/middleware"
	"github.com/gofiber/fiber/v2"
)

type GlobalErrorHandlerResp struct {
	Message string `json:"message"`
}

func main() {
	// Calling initializers
	initializers.StartEnvironmentVariables()

	validate := initializers.NewValidator()
	db := initializers.NewDatabaseConn()
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
			if fe, ok := err.(*fiber.Error); ok {
				return c.Status(fe.Code).JSON(GlobalErrorHandlerResp{
					Message: fe.Message,
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(GlobalErrorHandlerResp{
				Message: err.Error(),
			})
		},
	}
	app := fiber.New(fiberConfig)

	// Controllers
	userController := controllers.User{
		DB:       db,
		Validate: validate,
	}

	sessionController := controllers.Session{
		DB:       db,
		Validate: validate,
	}

	actorController := controllers.Actor{
		DB:       db,
		Validate: validate,
	}

	// Routes - Session
	app.Post("/login", sessionController.HandleLogin)
	app.Post("/logout", sessionController.HandleLogout)

	// Routes - User
	app.Post("/users", userController.CreateUser)
	app.Get("/users", middleware.VerifyAdmin, userController.ListAllUsersInDB)
	app.Get("/users/:uuid", middleware.VerifyUserOrAdmin, userController.GetUser)
	app.Delete("/users/:uuid", middleware.VerifyUserOrAdmin, userController.DeleteUser)
	app.Patch("/users/:uuid", middleware.VerifyUserOrAdmin, userController.UpdateUser)

	// Routes - Actor
	app.Post("/actors", middleware.VerifyAdmin, actorController.CreateActor)

	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
