package controllers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func (a *Actor) ListAllActorsInDB(c *fiber.Ctx) error {
	c.Accepts("application/json")

	// Query params
	offset := c.Query("offset", "0")
	limit := c.Query("limit", "10")
	orderBy := c.Query("sort", "created,desc")
	deletedQuery := c.Query("deleted", "false")

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		log.Println("Invalid offset value:", offset)
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Offset needs to be a valid integer",
		}
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		log.Println("Invalid limit value:", limit)
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Limit needs to be a valid integer",
		}
	}

	switch strings.ToLower(orderBy) {
	case "created,asc":
		orderBy = "created_at ASC"
	case "created,desc":
		orderBy = "created_at DESC"
	case "name,asc":
		orderBy = "name ASC"
	case "name,desc":
		orderBy = "name DESC"
	case "surname,asc":
		orderBy = "surname ASC"
	case "surname,desc":
		orderBy = "surname DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	var deleted bool
	if deletedQuery == "true" {
		deleted = true
	}

	actorsList, err := ActorModel.GetAllActors(a.DB, offsetInt, limitInt, orderBy, deleted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting all actors:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	c.Status(fiber.StatusOK).JSON(actorsList)
	return nil
}

func (a *Actor) GetActor(c *fiber.Ctx) error {
	c.Accepts("application/json")
	uuidParam := c.Params("uuid")

	uuid, err := uuid.Parse(uuidParam)
	if err != nil {
		log.Println("Invalid uuid sent in param:", err)
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid uuid parameter",
		}
	}

	actorResponse, err := ActorModel.GetActorById(a.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Actor id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Actor id not found in database",
			}
		}

		log.Println("Error getting actor by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusOK).JSON(actorResponse)
	return nil
}

// @TODO: Make controller method to get actor with movies
