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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
	app.Use(logger.New(logger.Config{
		Format: "IP+PORT: ${ip}:${port} | METHOD: ${method} | STATUS: ${status} | PATH: ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // This wouldn't be in a real application, just for testing purposes
	}))
	app.Use(recover.New())

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

	movieController := controllers.Movie{
		DB:       db,
		Validate: validate,
	}

	commentController := controllers.Comment{
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
	app.Get("/users/:uuid/comments", middleware.VerifyUserOrAdmin, userController.GetUserComments)
	app.Delete("/users/:uuid", middleware.VerifyUserOrAdmin, userController.DeleteUser)
	app.Patch("/users/:uuid", middleware.VerifyUserOrAdmin, userController.UpdateUser)

	// Routes - Actor
	app.Post("/actors", middleware.VerifyAdmin, actorController.CreateActor)
	app.Get("/actors", actorController.ListAllActorsInDB)
	app.Get("/actors/:uuid", actorController.GetActor)
	app.Get("/actors/:uuid/movies", actorController.GetActorMovies)
	app.Delete("/actors/:uuid", middleware.VerifyAdmin, actorController.DeleteActor)
	app.Patch("/actors/:uuid", middleware.VerifyAdmin, actorController.UpdateActor)

	// Routes - Movie
	app.Post("/movies", middleware.VerifyAdmin, movieController.CreateMovie)
	app.Post("/movies/:uuid/actors", middleware.VerifyAdmin, movieController.CreateActorsRelationshipsWithMovie)
	app.Get("/movies", movieController.ListAllMoviesInDB)
	app.Get("/movies/:uuid", movieController.GetMovie)
	app.Get("/movies/:uuid/comments", movieController.GetMovieComments)
	app.Delete("/movies/:uuid", middleware.VerifyAdmin, movieController.DeleteMovie)
	app.Delete("/movies/:uuid/actors", middleware.VerifyAdmin, movieController.DeleteActorsRelationshipsWithMovie)
	app.Patch("/movies/:uuid", middleware.VerifyAdmin, movieController.UpdateMovie)

	// Routes - Comments
	app.Post("/comments/:uuid", middleware.VerifyUserOrAdmin, commentController.CreateComment)
	app.Get("/comments", middleware.VerifyAdmin, commentController.ListAllCommentsInDb)
	app.Get("/comments/:uuid", commentController.GetComment)
	app.Delete("/comments/:uuid", middleware.VerifyUserOrAdmin, commentController.DeleteComment)
	app.Patch("/comments/:uuid", middleware.VerifyUserOrAdmin, commentController.UpdateComment)

	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
