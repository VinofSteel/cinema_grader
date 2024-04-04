package controllers

import (
	"database/sql"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/gofiber/fiber/v2"
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
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": err,
		})
		return err
	}

	user, err := UserModel.InsertUserInDB(u.DB, userBody)
	if err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"message": err,
		})
		return err
	}

	c.Status(fiber.StatusOK).JSON(user)
	return nil
}
