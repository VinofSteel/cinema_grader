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
	"golang.org/x/crypto/bcrypt"
)

// Controller type
type User struct {
	DB       *sql.DB
	Validate *validator.Validate
}

// User model
var UserModel models.UserModel

func (u *User) CreateUser(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var userBody models.UserBody
	if err := c.BodyParser(&userBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error while parsing JSON body",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, u.Validate, userBody); !valid {
		return nil
	}

	// Checking if user already exists in DB
	existingUser, err := UserModel.GetUserByEmail(u.DB, userBody.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting user by email:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	if existingUser.ID != uuid.Nil {
		log.Println("Trying to create user with existing email in db")
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "User with this email already exists",
		}
	}

	// Encrypting user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userBody.Password), 12)
	if err != nil {
		log.Println("Error encrypting user's password:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}
	userBody.Password = string(hashedPassword)

	user, err := UserModel.InsertUserInDB(u.DB, userBody)
	if err != nil {
		log.Println("Error inserting user in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(user)
	return nil
}

func (u *User) GetAllUsers(c *fiber.Ctx) error {
	c.Accepts("application/json")

	// Query params
	offset := c.Query("offset", "0")
	limit := c.Query("limit", "10")
	orderBy := c.Query("sort", "created,desc")

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
	case "email,asc":
		orderBy = "email ASC"
	case "email,desc":
		orderBy = "email DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	usersList, err := UserModel.GetAllUsers(u.DB, offsetInt, limitInt, orderBy)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting user by email:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	c.Status(fiber.StatusOK).JSON(usersList)
	return nil
}

func (u *User) GetUserById(c *fiber.Ctx) error {
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

	userInDb, err := UserModel.GetUserById(u.DB, uuid)
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

	c.Status(fiber.StatusOK).JSON(userInDb)
	return nil
}
