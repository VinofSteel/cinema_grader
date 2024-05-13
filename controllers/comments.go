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
type Comment struct {
	DB       *sql.DB
	Validate *validator.Validate
}

// Comment model
var CommentModel models.CommentModel

func (com *Comment) CreateComment(c *fiber.Ctx) error {
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

	userResponse, err := UserModel.GetUserById(com.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "User id not found in database",
			}
		}

		log.Println("Error getting user by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	if userResponse.DeletedAt.Valid {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Trying to comment as a deleted user, check your request",
		}
	}

	var commentBody models.CommentBody
	if err := c.BodyParser(&commentBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, com.Validate, commentBody); !valid {
		return nil
	}

	commentResponse, err := CommentModel.InsertCommentInDB(com.DB, uuid, commentBody)
	if err != nil {
		log.Println("Error inserting comment in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(commentResponse)
	return nil
}

func (com *Comment) ListAllCommentsInDb(c *fiber.Ctx) error {
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
	case "grade,asc":
		orderBy = "grade ASC"
	case "grade,desc":
		orderBy = "grade DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	var deleted bool
	if deletedQuery == "true" {
		deleted = true
	}

	commentsList, err := CommentModel.GetAllComments(com.DB, offsetInt, limitInt, orderBy, deleted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting all comments:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	c.Status(fiber.StatusOK).JSON(commentsList)
	return nil
}
