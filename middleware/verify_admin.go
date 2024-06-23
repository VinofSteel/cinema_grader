package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/VinOfSteel/cinemagrader/controllers"
	"github.com/gofiber/fiber/v2"
)

func VerifyAdmin(c *fiber.Ctx) error {
	sessionController := controllers.Session{}
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Missing Authorization header",
		}
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "Invalid Authorization header format",
		}
	}
	tokenString := parts[1]

	claims, err := sessionController.VerifyToken(tokenString)
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
