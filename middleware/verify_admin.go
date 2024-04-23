package middleware

import (
	"log"
	"time"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/gofiber/fiber/v2"
)

func VerifyAdmin(c *fiber.Ctx) error {
	sessionController := controllers.Session{}
	authCookie := c.Cookies("Authorization")

	claims, err := sessionController.VerifyToken(authCookie)
	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid or non-existing token",
		}
	}

	expirationTime := time.Unix(int64(claims["expiration"].(float64)), 0)
	if expirationTime.Before(time.Now()) {
		log.Println("Token has expired")
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Token has expired, login again",
		}
	}

	if isAdmin, ok := claims["isAdm"].(bool); !isAdmin || !ok {
		log.Printf("User with id %s and email %s tried to acess an admin only route.\n", claims["id"].(string), claims["email"].(string))
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Only administrators can access this route",
		}
	}

	return c.Next()
}
