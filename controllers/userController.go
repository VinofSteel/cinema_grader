package controllers

import (
	"database/sql"
	"log"

	"github.com/VinOfSteel/cinemagrader/models"
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
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
		return err
	}

	// Encrypting user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userBody.Password), 12)
	if err != nil {
		log.Println("Error encrypting user's password:", err)
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
	}
	userBody.Password = string(hashedPassword)

	user, err := UserModel.InsertUserInDB(u.DB, userBody)
	if err != nil {
		log.Println("Error inserting user in DB:", err)
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": "Internal Server Error",
		})
		return err
	}

	c.Status(fiber.StatusOK).JSON(user)
	return nil
}
