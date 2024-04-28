package controllers

import (
	"database/sql"
	"log"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Controller type
type Actor struct {
	DB       *sql.DB
	Validate *validator.Validate
}

// Actor model
var ActorModel models.ActorModel

func (a *Actor) CreateActor(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var actorBody models.ActorBody
	if err := c.BodyParser(&actorBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error while parsing JSON body",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, a.Validate, actorBody); !valid {
		return nil
	}

	actorResponse, err := ActorModel.InsertActorInDB(a.DB, actorBody)
	if err != nil {
		log.Println("Error inserting actor in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(actorResponse)
	return nil
}
