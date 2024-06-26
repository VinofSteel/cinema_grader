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
)

type Movie struct {
	DB       *sql.DB
	Validate *validator.Validate
}

var MovieModel models.MovieModel

func (m *Movie) CreateMovie(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var movieBody models.MovieBody
	if err := c.BodyParser(&movieBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, m.Validate, movieBody); !valid {
		return nil
	}

	existingMovie, err := MovieModel.GetMovieByTitle(m.DB, movieBody.Title)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting movie by title:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	if existingMovie.ID != uuid.Nil {
		log.Println("Trying to create a movie with duplicate title in DB")
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Movie with this title already exists",
		}
	}

	movieResponse, err := MovieModel.InsertMovieInDB(m.DB, movieBody)
	if err != nil {
		log.Println("Error inserting movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusCreated).JSON(movieResponse)
	return nil
}

func (m *Movie) ListAllMoviesInDB(c *fiber.Ctx) error {
	c.Accepts("application/json")

	offset := c.Query("offset", "0")
	limit := c.Query("limit", "10")
	orderBy := c.Query("sort", "created,desc")
	deletedQuery := c.Query("deleted", "false")
	actorsQuery := c.Query("with_actors", "false")

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
	case "title,asc":
		orderBy = "title ASC"
	case "title,desc":
		orderBy = "title DESC"
	case "director,asc":
		orderBy = "director ASC"
	case "director,desc":
		orderBy = "director DESC"
	case "release_date,asc":
		orderBy = "release_date ASC"
	case "release_date,desc":
		orderBy = "release_date DESC"
	case "average_grade,asc":
		orderBy = "average_grade ASC"
	case "average_grade,desc":
		orderBy = "average_grade DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	var deleted bool
	if deletedQuery == "true" {
		deleted = true
	}

	var withActors bool
	if actorsQuery == "true" {
		withActors = true
	}

	if withActors {
		moviesList, err := MovieModel.GetAllMoviesWithActors(m.DB, offsetInt, limitInt, orderBy, deleted)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Println("Error getting all movies with actors:", err)
				return &fiber.Error{
					Code:    fiber.StatusInternalServerError,
					Message: "Unknown error",
				}
			}
		}

		c.Status(fiber.StatusOK).JSON(moviesList)
		return nil
	}

	moviesList, err := MovieModel.GetAllMovies(m.DB, offsetInt, limitInt, orderBy, deleted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error getting all movies:", err)
			return &fiber.Error{
				Code:    fiber.StatusInternalServerError,
				Message: "Unknown error",
			}
		}
	}

	c.Status(fiber.StatusOK).JSON(moviesList)
	return nil
}

func (m *Movie) GetMovie(c *fiber.Ctx) error {
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

	movieResponse, err := MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusOK).JSON(movieResponse)
	return nil
}

func (m *Movie) DeleteMovie(c *fiber.Ctx) error {
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

	_, err = MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	if err := MovieModel.DeleteMovieById(m.DB, uuid); err != nil {
		log.Println("Error deleting movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Couldn't delete movie in DB",
		}
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (m *Movie) UpdateMovie(c *fiber.Ctx) error {
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

	_, err = MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	var movieBody models.MovieEditBody
	if err := c.BodyParser(&movieBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, m.Validate, movieBody); !valid {
		return nil
	}

	// Verifying that the title is not a duplicate
	existingMovie, err := MovieModel.GetMovieByTitle(m.DB, movieBody.Title)
	if err == nil && existingMovie.ID != uuid {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Movie with this title already exists",
		}
	}

	movieResponse, err := MovieModel.UpdateMovieById(m.DB, uuid, movieBody)
	if err != nil {
		log.Println("Error updating movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusOK).JSON(movieResponse)
	return nil
}

func (m *Movie) CreateActorsRelationshipsWithMovie(c *fiber.Ctx) error {
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

	movieResponse, err := MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	if movieResponse.DeletedAt.Valid {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "Trying to create an actor relationship on a deleted movie, check your request",
		}
	}

	var movieActorsBody models.MovieActorsBody
	if err := c.BodyParser(&movieActorsBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, m.Validate, movieActorsBody); !valid {
		return nil
	}

	// Checking if trying to add an actor already in the movie to the movie
	actorUUIDs := make(map[string]struct{})
	for _, actorUUID := range movieResponse.Actors {
		actorUUIDs[actorUUID.ID.String()] = struct{}{}
	}

	for _, actorIDInBody := range movieActorsBody.Actors {
		if _, ok := actorUUIDs[actorIDInBody]; ok {
			return &fiber.Error{
				Code:    fiber.StatusBadRequest,
				Message: "There is an actor already in the movie on the request",
			}
		}
	}

	if err := MovieModel.InsertActorsRelationshipsWithMovie(m.DB, uuid, movieActorsBody); err != nil {
		log.Println("Error associating actors with movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Couldn't associate actors with movie, check your request",
		}
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (m *Movie) DeleteActorsRelationshipsWithMovie(c *fiber.Ctx) error {
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

	movieResponse, err := MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	var movieActorsBody models.MovieActorsBody
	if err := c.BodyParser(&movieActorsBody); err != nil {
		log.Println("Error parsing JSON body:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Error while parsing JSON body, check your request",
		}
	}

	// Validating input data. We return "nil" because the ValidateData function sends a response back by itself and we need to return here to stop the function.
	if valid := validation.ValidateData(c, m.Validate, movieActorsBody); !valid {
		return nil
	}

	// Checking if trying to delete an actor that is not on the movie
	actorUUIDs := make(map[string]struct{})
	for _, actorUUID := range movieResponse.Actors {
		actorUUIDs[actorUUID.ID.String()] = struct{}{}
	}

	for _, actorIDInBody := range movieActorsBody.Actors {
		if _, ok := actorUUIDs[actorIDInBody]; !ok {
			return &fiber.Error{
				Code:    fiber.StatusBadRequest,
				Message: "Trying to delete an actor that is already not on the movie",
			}
		}
	}

	if err := MovieModel.DeleteActorsRelationshipsWithMovie(m.DB, uuid, movieActorsBody); err != nil {
		log.Println("Error deleting actors from movie in DB:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Couldn't delete actors from movie, check your request",
		}
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (m *Movie) GetMovieComments(c *fiber.Ctx) error {
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

	// Query params
	orderBy := c.Query("sort", "created,desc")
	deletedQuery := c.Query("deleted", "false")

	switch strings.ToLower(orderBy) {
	case "created,asc":
		orderBy = "created_at ASC"
	case "created,desc":
		orderBy = "created_at DESC"
	case "grade,asc":
		orderBy = "grade ASC"
	case "grade,desc":
		orderBy = "grade DESC"
	case "updated,asc":
		orderBy = "updated_at ASC"
	default:
		orderBy = "updated_at DESC"
	}

	var deleted bool
	if deletedQuery == "true" {
		deleted = true
	}

	_, err = MovieModel.GetMovieByIdWithActors(m.DB, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	movieWithCommentsResponse, err := CommentModel.GetAllCommentsInAMovieInDb(m.DB, uuid, orderBy, deleted)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Movie id not found in database:", err)
			return &fiber.Error{
				Code:    fiber.StatusNotFound,
				Message: "Movie id not found in database",
			}
		}

		log.Println("Error getting movie by id:", err)
		return &fiber.Error{
			Code:    fiber.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	c.Status(fiber.StatusOK).JSON(movieWithCommentsResponse)
	return nil
}
