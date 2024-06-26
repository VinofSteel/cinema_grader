package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/VinOfSteel/cinemagrader/validation"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Controller type
type Session struct {
	DB       *sql.DB
	Validate *validator.Validate
}

// Login types
type LoginBody struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type LoginResponse struct {
	UserID uuid.UUID `json:"userId"`
	Token  string    `json:"token"`
}

func createToken(uuid uuid.UUID, email string, isAdm bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         uuid,
		"email":      email,
		"isAdm":      isAdm,
		"expiration": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Session) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *Session) HandleLogin(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var loginData LoginBody
	if err := c.BodyParser(&loginData); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	if valid := validation.ValidateData(c, s.Validate, loginData); !valid {
		return nil
	}

	// Verifying if user exists in DB
	existingUser, err := UserModel.GetUserByEmail(s.DB, loginData.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting user by email:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	if existingUser.ID == uuid.Nil {
		log.Println("Trying to login with an email that does not exist in DB")
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid email/password",
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(loginData.Password)); err != nil {
		log.Println("Password does not match:", err)
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Invalid email/password",
		}
	}

	token, err := createToken(existingUser.ID, existingUser.Email, existingUser.IsAdm)
	if err != nil {
		log.Println("Couldn't create JWT:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Couldn't create user token",
		}
	}

	loginResponse := LoginResponse{
		UserID: existingUser.ID,
		Token:  token,
	}
	c.Status(fiber.StatusOK).JSON(loginResponse)

	return nil
}
