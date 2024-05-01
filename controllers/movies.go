package controllers

import (
	"database/sql"
	"log"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Movie struct {
	DB       *sql.DB
	Validate *validator.Validate
}

var MovieModel models.MovieModel

func (m *Movie) CreateMovie(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var movieBody models.MovieBody
	if err := c.BodyParser(&movieBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error while parsing JSON body",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, m.Validate, movieBody); !valid {
		return nil
	}

	// @TODO: Check if movie already exists in DB
	existingMovie, err := MovieModel.GetMovieByTitle(m.DB, movieBody.Title)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error movie user by title:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	if existingMovie.ID != uuid.Nil {
		log.Println("Trying to create a movie with duplicate title in DB")
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Movie with this title already exists",
		}
	}

	movieResponse, err := MovieModel.InsertMovieInDB(m.DB, movieBody)
	if err != nil {
		log.Println("Error inserting movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(movieResponse)
	return nil
}
