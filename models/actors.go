package models

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ActorModel struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Surname   string       `json:"surname"`
	Birthday  string       `json:"birthday"`
	Picture   string       `json:"picture"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	CreatorId string `json:"creatorId"`
}

type ActorBody struct {
	Name     string `json:"name" validate:"required"`
	Surname  string `json:"surname" validate:"required"`
	Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
	Picture  string `json:"picture" validate:"omitempty"`

	CreatorId string `json:"creatorId" validate:"required,isadminuuid"`
}

type ActorEditBody struct {
	Name     string `json:"name" validate:"omitempty"`
	Surname  string `json:"surname" validate:"omitempty"`
	Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
	Picture  string `json:"picture" validate:"omitempty"`
}

type ActorResponse struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Surname   string       `json:"surname"`
	Birthday  string       `json:"birthday"`
	Picture   string       `json:"picture"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	CreatorId string `json:"creatorId"`
}

type ActorResponseWithMovies struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Surname   string       `json:"surname"`
	Birthday  string       `json:"birthday"`
	Picture   string       `json:"picture"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	CreatorId string          `json:"creatorId"`
	Movies    []MovieResponse `json:"movies"`
}

func (a *ActorModel) InsertActorInDB(db *sql.DB, actorInfo ActorBody) (ActorResponse, error) {
	log.Printf("Inserting actor with name %s in DB by user %s...\n", actorInfo.Name, actorInfo.CreatorId)

	query := `INSERT INTO actors
			(name, surname, birthday, picture, creator_id)
			VALUES ($1, $2, $3, $4, $5)
				RETURNING id, name, surname, birthday, picture, created_at, updated_at, deleted_at, creator_id;`

	var actor ActorResponse

	if err := db.QueryRow(query, actorInfo.Name, actorInfo.Surname, actorInfo.Birthday, actorInfo.Picture, actorInfo.CreatorId).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
		log.Printf("Error inserting actor into database: %v\n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}

func (a *ActorModel) GetAllActors(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]ActorResponse, error) {
	log.Printf("Getting all actors in DB, with offset %v, limit %v, orderBy %v and deleted %v...\n", offset, limit, orderBy, deleted)

	var getActorsQueryBuilder strings.Builder
	getActorsQueryBuilder.WriteString(`SELECT 
	id, name, surname, birthday, picture, created_at, updated_at, deleted_at, creator_id 
	FROM actors`)

	if !deleted {
		getActorsQueryBuilder.WriteString(" WHERE deleted_at IS NULL")
	}

	getActorsQueryBuilder.WriteString(" ORDER BY " + orderBy + " OFFSET $1 LIMIT $2;")

	query := getActorsQueryBuilder.String()
	rows, err := db.Query(query, offset, limit)
	if err != nil {
		log.Println("Error getting all actors from db:", err)
		return nil, err
	}
	defer rows.Close()

	var actors []ActorResponse
	for rows.Next() {
		var actor ActorResponse
		if err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
			log.Println("Error scanning actor from db:", err)
			return nil, err
		}
		actors = append(actors, actor)
	}

	return actors, nil
}

func (a *ActorModel) GetActorById(db *sql.DB, uuid uuid.UUID) (ActorResponse, error) {
	log.Printf("Getting actor with uuid %s in DB... \n", uuid)

	query := `SELECT 
		id, name, surname, birthday, picture, created_at, updated_at, deleted_at, creator_id
        FROM actors
        	WHERE id = $1;`

	var actor ActorResponse
	if err := db.QueryRow(query, uuid).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
		log.Printf("Error getting actor by id in the database: %v\n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}

func (a *ActorModel) GetActorByIdWithMovies(db *sql.DB, uuid uuid.UUID) (ActorResponseWithMovies, error) {
	log.Printf("Getting actor with uuid %s in DB with movies... \n", uuid)

	// Verifying if uuid actually exists in the DB before proceeding with the query
	_, err := a.GetActorById(db, uuid)
	if err != nil {
		log.Printf("Error getting actor by id in the database: %v\n", err)
		return ActorResponseWithMovies{}, err
	}

	query := `SELECT 
		a.id, a.name, a.surname, a.birthday, a.picture,
		a.created_at, a.updated_at, a.deleted_at, 
		a.creator_id, 
		m.id, m.title, m.director, m.release_date, 
		m.average_grade, m.picture,
		m.created_at, m.updated_at, m.deleted_at,
		m.creator_id 
			FROM actors a
				LEFT JOIN movies_actors ma ON a.id = ma.actor_id
				LEFT JOIN movies m ON ma.movie_id = m.id
					WHERE a.id = $1;`

	var actor ActorResponseWithMovies
	rows, err := db.Query(query, uuid)
	if err != nil {
		log.Printf("Error getting actor by id with movies from database: %v\n", err)
		return ActorResponseWithMovies{}, err
	}
	defer rows.Close()

	movies := make([]MovieResponse, 0)
	for rows.Next() {
		var movie MovieResponse
		err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId, &movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.Picture, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
		if err != nil {
			log.Printf("Error scanning movie row in GetActorByIdWithMovies: %v\n", err)
			continue
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows in GetActorByIdWithMovies: %v\n", err)
		return ActorResponseWithMovies{}, err
	}

	actor.Movies = movies

	return actor, nil
}

func (a *ActorModel) DeleteActorById(db *sql.DB, uuid uuid.UUID) error {
	log.Printf("Deleting actor with uuid %s in DB... \n", uuid)

	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error beginning transaction made while deleting actor by id: %v\n", err)
		return err
	}

	defer func() {
		if err != nil {
			log.Printf("Rolling back transaction made while deleting actor by id due to error: %v\n", err)
			tx.Rollback()
			return
		}
	}()

	// Delete entries from the pivot table if they exist
	deleteMoviesQuery := `DELETE FROM movies_actors WHERE actor_id = $1;`

	_, err = tx.Exec(deleteMoviesQuery, uuid)
	if err != nil {
		log.Printf("Error deleting actor's associations with movies while deleting actor by id: %v\n", err)
		return err
	}

	// Deleting actor per se
	query := `UPDATE actors 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL;`

	_, err = tx.Exec(query, uuid)
	if err != nil {
		log.Printf("Error deleting actor by uuid: %v\n", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction made while deleting actor by id: %v\n", err)
		return err
	}

	return nil
}

func (a *ActorModel) UpdateActorById(db *sql.DB, uuid uuid.UUID, body ActorEditBody) (ActorResponse, error) {
	log.Printf("Updating actor with uuid %s in DB... \n", uuid)

	var updateQueryBuilder strings.Builder
	var args []interface{}

	updateQueryBuilder.WriteString("UPDATE actors SET ")

	argIndex := 1
	if body.Name != "" {
		updateQueryBuilder.WriteString("name = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Name)
		argIndex++
	}

	if body.Surname != "" {
		updateQueryBuilder.WriteString("surname = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Surname)
		argIndex++
	}

	if body.Birthday != "" {
		updateQueryBuilder.WriteString("birthday = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Birthday)
		argIndex++
	}

	if body.Picture != "" {
		updateQueryBuilder.WriteString("picture = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Picture)
		argIndex++
	}

	updateQueryBuilder.WriteString("updated_at = CURRENT_TIMESTAMP, ")
	query := strings.TrimSuffix(updateQueryBuilder.String(), ", ")
	query += " WHERE id = $" + strconv.Itoa(argIndex) + " AND deleted_at IS NULL RETURNING id, name, surname, birthday, picture, created_at, updated_at, deleted_at, creator_id;"
	args = append(args, uuid)

	var actor ActorResponse
	if err := db.QueryRow(query, args...).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.Picture, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
		log.Printf("Error updating actor by uuid: %v \n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}
