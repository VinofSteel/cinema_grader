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
			Message: "Error while parsing JSON body, check your request",
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
		log.Println("Trying to create user with existing email in DB")
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

	userResponse, err := UserModel.InsertUserInDB(u.DB, userBody)
	if err != nil {
		log.Println("Error inserting user in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(userResponse)
	return nil
}

func (u *User) ListAllUsersInDB(c *fiber.Ctx) error {
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
	case "email,asc":
		orderBy = "email ASC"
	case "email,desc":
		orderBy = "email DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	var deleted bool
	if deletedQuery == "true" {
		deleted = true
	}

	usersList, err := UserModel.GetAllUsers(u.DB, offsetInt, limitInt, orderBy, deleted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting all users:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	c.Status(fiber.StatusOK).JSON(usersList)
	return nil
}

func (u *User) GetUser(c *fiber.Ctx) error {
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

	userResponse, err := UserModel.GetUserById(u.DB, uuid)
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

	c.Status(fiber.StatusOK).JSON(userResponse)
	return nil
}

func (u *User) DeleteUser(c *fiber.Ctx) error {
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

	_, err = UserModel.GetUserById(u.DB, uuid)
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

	if err := UserModel.DeleteUserById(u.DB, uuid); err != nil {
		log.Println("Error deleting user in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Couldn't delete user in DB",
		}
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (u *User) UpdateUser(c *fiber.Ctx) error {
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

	_, err = UserModel.GetUserById(u.DB, uuid)
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

	var userBody models.UserEditBody
	if err := c.BodyParser(&userBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, u.Validate, userBody); !valid {
		return nil
	}

	userResponse, err := UserModel.UpdateUserById(u.DB, uuid, userBody)
	if err != nil {
		log.Println("Error updating user in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusOK).JSON(userResponse)
	return nil
}
