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
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
		return err
	}

	if errors := validation.ValidateData(userBody); len(errors) > 0 && errors[0].Error {
		errMsgs := make([]string, 0)

		for _, err := range errors {
			errMsgs = append(errMsgs, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: strings.Join(errMsgs, ", "),
		}
	}

	// Encrypting user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userBody.Password), 12)
	if err != nil {
		log.Println("Error encrypting user's password:", err)
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
	}
	userBody.Password = string(hashedPassword)

	user, err := UserModel.InsertUserInDB(u.DB, userBody)
	if err != nil {
		log.Println("Error inserting user in DB:", err)
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
		return err
	}

	c.Status(fiber.StatusOK).JSON(user)
	return nil
}
