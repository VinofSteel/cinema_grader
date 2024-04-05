package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Controller type
type User struct {
	DB *sql.DB
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

	if errors := validation.ValidateData(userBody); len(errors) > 0 && errors[0].Error {
		errMsgs := make([]string, 0)

		for _, err := range errors {
			errMsgs = append(errMsgs, fmt.Sprintf(
				"Field %s: '%s'",
				err.FailedField,
				err.ErrorMessage,
			))
		}

		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: strings.Join(errMsgs, " | "),
		}
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

	if existingUser.ID != "" {
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

	c.Status(fiber.StatusOK).JSON(user)
	return nil
}
