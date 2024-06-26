package models

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MovieModel struct {
	ID           uuid.UUID    `json:"id"`
	Title        string       `json:"title"`
	Director     string       `json:"director"`
	ReleaseDate  string       `json:"releaseDate"`
	AverageGrade float64      `json:"averageGrade"`
	Picture      string       `json:"picture"`
	Synopsis     string       `json:"synopsis"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DeletedAt    sql.NullTime `json:"deletedAt"`

	CreatorId string         `json:"creatorId"`
	Actors    []ActorModel   `json:"actors"`
	Comments  []CommentModel `json:"comments"`
}

type MovieBody struct {
	Title       string `json:"title" validate:"required"`
	Director    string `json:"director" validate:"required"`
	ReleaseDate string `json:"releaseDate" validate:"required,datetime=2006-01-02"`
	Picture     string `json:"picture" validate:"omitempty"`
	Synopsis    string `json:"synopsis" validate:"omitempty"`

	CreatorId string   `json:"creatorId" validate:"required,isadminuuid"`
	Actors    []string `json:"actors" validate:"required,unique,validactorslice"`
}

type MovieEditBody struct {
	Title       string `json:"title" validate:"omitempty"`
	Director    string `json:"director" validate:"omitempty"`
	ReleaseDate string `json:"releaseDate" validate:"omitempty,datetime=2006-01-02"`
	Picture     string `json:"picture" validate:"omitempty"`
	Synopsis    string `json:"synopsis" validate:"omitempty"`
}

type MovieActorsBody struct {
	Actors []string `json:"actors" validate:"required,unique,validactorslice"`
}

type MovieResponse struct {
	ID           uuid.UUID    `json:"id"`
	Title        string       `json:"title"`
	Director     string       `json:"director"`
	ReleaseDate  string       `json:"releaseDate"`
	AverageGrade float64      `json:"averageGrade"`
	Picture      string       `json:"picture"`
	Synopsis     string       `json:"synopsis"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DeletedAt    sql.NullTime `json:"deletedAt"`

	CreatorId string `json:"creatorId"`
}

type MovieResponseWithActors struct {
	ID           uuid.UUID    `json:"id"`
	Title        string       `json:"title"`
	Director     string       `json:"director"`
	ReleaseDate  string       `json:"releaseDate"`
	AverageGrade float64      `json:"averageGrade"`
	Picture      string       `json:"picture"`
	Synopsis     string       `json:"synopsis"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DeletedAt    sql.NullTime `json:"deletedAt"`

	CreatorId string          `json:"creatorId"`
	Actors    []ActorResponse `json:"actors"`
}

type MovieResponseWithActorsWithComments struct {
	MovieResponseWithActors
	Comments []CommentResponse
}

var actorModel ActorModel

// Internal methods
func (m *MovieModel) getActorsOfAMovie(db *sql.DB, movieID uuid.UUID) ([]ActorResponse, error) {
	query := `SELECT 
        a.id, a.name, a.surname, a.birthday, a.picture, a.created_at, a.updated_at, a.deleted_at, a.creator_id
        FROM actors a
        	JOIN movies_actors ma ON a.id = ma.actor_id
        		WHERE ma.movie_id = $1 AND a.deleted_at IS NULL;`

	rows, err := db.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actors []ActorResponse
	for rows.Next() {
		var actor ActorResponse
		if err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}

	return actors, nil
}

// Public methods
func (m *MovieModel) InsertMovieInDB(db *sql.DB, movieInfo MovieBody) (MovieResponseWithActors, error) {
	log.Printf("Inserting movie with title %s in DB by user %s...\n", movieInfo.Title, movieInfo.CreatorId)

	// Starting a transaction that can be rolled back if shit happens
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction to insert movie in DB: %v\n", err)
		return MovieResponseWithActors{}, err
	}
	defer tx.Rollback() // This will execute if any error is returned and cancel any changes to the db

	query := `INSERT INTO movies
			(title, director, release_date, picture, synopsis, creator_id)
			VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id;`

	var movie MovieResponseWithActors
	err = tx.QueryRow(query, movieInfo.Title, movieInfo.Director, movieInfo.ReleaseDate, movieInfo.Picture, movieInfo.Synopsis, movieInfo.CreatorId).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
	if err != nil {
		log.Printf("Error inserting movie into database: %v\n", err)
		return MovieResponseWithActors{}, err
	}

	// Associate actors with the movie in the pivot table
	var wg sync.WaitGroup
	errCh := make(chan error, len(movieInfo.Actors))
	actorInfoCh := make(chan ActorResponse, len(movieInfo.Actors))
	actorIndexMap := make(map[uuid.UUID]int)
	for i, actorID := range movieInfo.Actors {
		actorUUID, err := uuid.Parse(actorID)
		if err != nil {
			log.Printf("Error parsing actor id into uuid: %v\n", err)
			return MovieResponseWithActors{}, err
		}

		wg.Add(1)
		actorIndexMap[actorUUID] = i
		go func(actorUUID uuid.UUID) {
			defer wg.Done()

			actorResponse, err := actorModel.GetActorById(db, actorUUID)
			if err != nil {
				log.Printf("Trying to associate non-existant actor %v to a movie: %v\n", actorUUID, err)
				errCh <- err
				return
			}

			if actorResponse.DeletedAt.Valid {
				errCh <- fmt.Errorf("trying to insert deleted actor with ID %v and name %v into movie with title %v", actorResponse.ID, actorResponse.Name, movieInfo.Title)
				return
			}
			actorInfoCh <- actorResponse

			query := `INSERT INTO movies_actors (actor_id, movie_id) VALUES ($1, $2)`
			_, err = tx.Exec(query, actorUUID, movie.ID)
			if err != nil {
				log.Printf("Error associating actor %v with movie: %v\n", actorUUID, err)
				errCh <- err
				return
			}
		}(actorUUID)
	}

	wg.Wait()
	close(errCh)
	close(actorInfoCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error in insertion goroutine: %v", err)
			tx.Rollback()
			return MovieResponseWithActors{}, err
		}
	}

	// Getting the actors of the movie to fill the Actors slice and sorting the returning slice
	var actorResponses []ActorResponse
	for actorResp := range actorInfoCh {
		actorResponses = append(actorResponses, actorResp)
	}

	sort.Slice(actorResponses, func(i, j int) bool {
		indexI := actorIndexMap[actorResponses[i].ID]
		indexJ := actorIndexMap[actorResponses[j].ID]

		return indexI < indexJ
	})

	movie.Actors = actorResponses

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction while inserting movies in DB: %v\n", err)
		return MovieResponseWithActors{}, err
	}

	return movie, nil
}

func (m *MovieModel) GetAllMovies(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]MovieResponse, error) {
	log.Printf("Getting all movies in DB, with offset %v, limit %v, orderBy %v, no actors and deleted %v...\n", offset, limit, orderBy, deleted)

	var getMoviesQueryBuilder strings.Builder
	getMoviesQueryBuilder.WriteString(`SELECT
	id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id 
	FROM movies`)

	if !deleted {
		getMoviesQueryBuilder.WriteString(" WHERE deleted_at IS NULL")
	}

	getMoviesQueryBuilder.WriteString(" ORDER BY " + orderBy + " OFFSET $1 LIMIT $2;")

	query := getMoviesQueryBuilder.String()
	rows, err := db.Query(query, offset, limit)
	if err != nil {
		log.Println("Error getting all movies from db without actors:", err)
		return nil, err
	}
	defer rows.Close()

	var movies []MovieResponse
	for rows.Next() {
		var movie MovieResponse
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, nil
}

func (m *MovieModel) GetAllMoviesWithActors(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]MovieResponseWithActors, error) {
	log.Printf("Getting all movies in DB, with offset %v, limit %v, orderBy %v, with actors and deleted %v...\n", offset, limit, orderBy, deleted)

	var getMoviesQueryBuilder strings.Builder
	getMoviesQueryBuilder.WriteString(`SELECT
	id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id 
	FROM movies`)

	if !deleted {
		getMoviesQueryBuilder.WriteString(" WHERE deleted_at IS NULL")
	}

	getMoviesQueryBuilder.WriteString(" ORDER BY " + orderBy + " OFFSET $1 LIMIT $2;")

	query := getMoviesQueryBuilder.String()
	rows, err := db.Query(query, offset, limit)
	if err != nil {
		log.Println("Error getting all movies from db without actors:", err)
		return nil, err
	}
	defer rows.Close()

	var movies []MovieResponseWithActors
	for rows.Next() {
		var movie MovieResponseWithActors
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId); err != nil {
			return nil, err
		}

		actors, err := m.getActorsOfAMovie(db, movie.ID)
		if err != nil {
			log.Printf("Error getting actors of movie %v, %v", movie.Title, err)
			return nil, err
		}
		movie.Actors = actors

		movies = append(movies, movie)
	}

	return movies, nil
}

func (m *MovieModel) GetMovieByTitle(db *sql.DB, title string) (MovieModel, error) {
	log.Printf("Getting movie with title %s in DB... \n", title)

	query := `SELECT 
		id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id 
		FROM movies 
			WHERE title = $1;`

	var movie MovieModel
	err := db.QueryRow(query, title).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
	if err != nil {
		log.Printf("Error getting movie by title: %v\n", err)
		return MovieModel{}, err
	}

	return movie, nil
}

func (m *MovieModel) GetMovieByIdWithActors(db *sql.DB, uuid uuid.UUID) (MovieResponseWithActors, error) {
	log.Printf("Getting movie with id %s in DB... \n", uuid)

	query := `SELECT 
		id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id 
		FROM movies 
			WHERE id = $1;`

	var movie MovieResponseWithActors
	err := db.QueryRow(query, uuid).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
	if err != nil {
		log.Printf("Error getting movie by id: %v\n", err)
		return MovieResponseWithActors{}, err
	}

	actors, err := m.getActorsOfAMovie(db, uuid)
	if err != nil {
		log.Printf("Error getting actors of movie %v, %v", movie.Title, err)
		return MovieResponseWithActors{}, err
	}
	movie.Actors = actors

	return movie, nil
}

func (m *MovieModel) DeleteMovieById(db *sql.DB, uuid uuid.UUID) error {
	log.Printf("Deleting movie with uuid %s in DB... \n", uuid)

	query := `UPDATE movies 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL;`

	_, err := db.Exec(query, uuid)
	if err != nil {
		log.Printf("Error deleting movie by uuid: %v\n", err)
		return err
	}

	return nil
}

func (m *MovieModel) UpdateMovieById(db *sql.DB, uuid uuid.UUID, body MovieEditBody) (MovieResponse, error) {
	log.Printf("Updating movie with uuid %s in DB... \n", uuid)

	var updateQueryBuilder strings.Builder
	var args []interface{}

	updateQueryBuilder.WriteString("UPDATE movies SET ")

	argIndex := 1
	if body.Title != "" {
		updateQueryBuilder.WriteString("title = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Title)
		argIndex++
	}

	if body.Director != "" {
		updateQueryBuilder.WriteString("director = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Director)
		argIndex++
	}

	if body.ReleaseDate != "" {
		updateQueryBuilder.WriteString("release_date = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.ReleaseDate)
		argIndex++
	}

	if body.Picture != "" {
		updateQueryBuilder.WriteString("picture = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Picture)
		argIndex++
	}

	if body.Synopsis != "" {
		updateQueryBuilder.WriteString("synopsis = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Synopsis)
		argIndex++
	}

	updateQueryBuilder.WriteString("updated_at = CURRENT_TIMESTAMP, ")
	query := strings.TrimSuffix(updateQueryBuilder.String(), ", ")
	query += " WHERE id = $" + strconv.Itoa(argIndex) + " AND deleted_at IS NULL RETURNING id, title, director, release_date, average_grade, picture, synopsis, created_at, updated_at, deleted_at, creator_id;"
	args = append(args, uuid)

	var movie MovieResponse
	if err := db.QueryRow(query, args...).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.Synopsis, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId); err != nil {
		log.Printf("Error updating movie by uuid: %v \n", err)
		return MovieResponse{}, nil
	}

	return movie, nil
}

func (m *MovieModel) InsertActorsRelationshipsWithMovie(db *sql.DB, id uuid.UUID, body MovieActorsBody) error {
	log.Printf("Associating actors to movie with uuid %s in DB... \n", id)

	var wg sync.WaitGroup
	errCh := make(chan error, len(body.Actors))
	actorInfoCh := make(chan ActorResponse, len(body.Actors))
	for _, actorID := range body.Actors {
		actorUUID, err := uuid.Parse(actorID)
		if err != nil {
			log.Printf("Error parsing actor id into uuid: %v\n", err)
			return err
		}

		wg.Add(1)
		go func(actorUUID uuid.UUID) {
			defer wg.Done()

			// This seems like error handling, but it's actually necessary because we
			// only have the uuid and we need the full info of the actor.
			actorResponse, err := actorModel.GetActorById(db, actorUUID)
			if err != nil {
				log.Printf("Trying to associate non-existant actor %v to a movie: %v\n", actorUUID, err)
				errCh <- err
				return
			}

			if actorResponse.DeletedAt.Valid {
				errCh <- fmt.Errorf("trying to insert deleted actor with ID %v and name %v into movie with id %v", actorResponse.ID, actorResponse.Name, id)
				return
			}
			actorInfoCh <- actorResponse

			query := `INSERT INTO movies_actors (actor_id, movie_id) VALUES ($1, $2);`
			_, err = db.Exec(query, actorUUID, id)
			if err != nil {
				log.Printf("Error associating actor %v with movie: %v\n", actorUUID, err)
				errCh <- err
				return
			}
		}(actorUUID)
	}

	wg.Wait()
	close(errCh)
	close(actorInfoCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error in insertion goroutine: %v", err)
			return err
		}
	}

	return nil
}

func (m *MovieModel) DeleteActorsRelationshipsWithMovie(db *sql.DB, id uuid.UUID, body MovieActorsBody) error {
	log.Printf("Deleting actors associated movie with uuid %s in DB... \n", id)

	var wg sync.WaitGroup
	errCh := make(chan error, len(body.Actors))
	actorInfoCh := make(chan ActorResponse, len(body.Actors))
	for _, actorID := range body.Actors {
		actorUUID, err := uuid.Parse(actorID)
		if err != nil {
			log.Printf("Error parsing actor id into uuid: %v\n", err)
			return err
		}

		wg.Add(1)
		go func(actorUUID uuid.UUID) {
			defer wg.Done()

			// This seems like error handling, but it's actually necessary because we
			// only have the uuid and we need the full info of the actor.
			actorResponse, err := actorModel.GetActorById(db, actorUUID)
			if err != nil {
				log.Printf("Trying to delete non-existant actor %v from movie with ID %v\n", actorUUID, err)
				errCh <- err
				return
			}

			if actorResponse.DeletedAt.Valid {
				errCh <- fmt.Errorf("trying to delete actor with ID %v and name %v from movie with ID %v", actorResponse.ID, actorResponse.Name, id)
				return
			}
			actorInfoCh <- actorResponse

			query := `DELETE FROM movies_actors WHERE actor_id = $1;`
			_, err = db.Exec(query, actorUUID)
			if err != nil {
				log.Printf("Error deleting actor %v from movie: %v\n", actorUUID, err)
				errCh <- err
				return
			}
		}(actorUUID)
	}

	wg.Wait()
	close(errCh)
	close(actorInfoCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error in insertion goroutine: %v", err)
			return err
		}
	}

	return nil
}
