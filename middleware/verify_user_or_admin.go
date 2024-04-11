package middleware

import (
	"log"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func VerifyUserOrAdmin(c *fiber.Ctx) error {
	sessionController := controllers.Session{}
	authCookie := c.Cookies("Authorization")
	queryId := c.Params("uuid")

	if _, err := uuid.Parse(queryId); err != nil {
		log.Println("Invalid uuid ent in param:", err)
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid uuid parameter",
		}
	}

	claims, err := sessionController.VerifyToken(authCookie)
	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid or non-existing token",
		}
	}

	isAdmin := claims["isAdm"].(bool)
	id := claims["id"].(string)

	if id != queryId && !isAdmin {
		log.Printf("User with id %s trying to modify user with id %s is not an admin.\n", id, queryId)
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "This route is only accessible to administrators or by the user with the same id as the parameter",
		}

	}

	return c.Next()
}
