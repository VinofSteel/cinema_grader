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
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	CreatorId string       `json:"creatorId"`
	Movies    []MovieModel `json:"movies"`
}

type ActorBody struct {
	Name      string `json:"name" validate:"required"`
	Surname   string `json:"surname" validate:"required"`
	Birthday  string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
	CreatorId string `json:"creatorId" validate:"required,isAdminUuid"`
}

type ActorEditBody struct {
	Name     string `json:"name" validate:"omitempty"`
	Surname  string `json:"surname" validate:"omitempty"`
	Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
}

type ActorResponse struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Surname   string       `json:"surname"`
	Birthday  string       `json:"birthday"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	CreatorId string `json:"creatorId"`
}

func (a *ActorModel) InsertActorInDB(db *sql.DB, actorInfo ActorBody) (ActorResponse, error) {
	log.Printf("Inserting actor with name %s in DB by user %s...\n", actorInfo.Name, actorInfo.CreatorId)

	query := `INSERT INTO actors
			(name, surname, birthday, creator_id)
			VALUES ($1, $2, $3, $4)
				RETURNING id, name, surname, birthday, created_at, updated_at, deleted_at, creator_id;`

	var actor ActorResponse
	err := db.QueryRow(query, actorInfo.Name, actorInfo.Surname, actorInfo.Birthday, actorInfo.CreatorId).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId)
	if err != nil {
		log.Printf("Error inserting actor into database: %v\n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}

func (a *ActorModel) GetAllActors(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]ActorResponse, error) {
	log.Printf("Getting all actors in DB, with offset %v, limit %v, orderBy %v and deleted %v...\n", offset, limit, orderBy, deleted)

	var getActorsQueryBuilder strings.Builder
	getActorsQueryBuilder.WriteString(`SELECT 
	id, name, surname, birthday, created_at, updated_at, deleted_at, creator_id 
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
		if err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId); err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}

	return actors, nil
}

func (a *ActorModel) GetActorById(db *sql.DB, uuid uuid.UUID) (ActorResponse, error) {
	log.Printf("Getting actor with uuid %s in DB... \n", uuid)

	query := `SELECT 
		id, name, surname, birthday, created_at, updated_at, deleted_at, creator_id
        FROM actors
        	WHERE id = $1;`

	var actor ActorResponse
	err := db.QueryRow(query, uuid).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId)
	if err != nil {
		log.Printf("Error getting actor by id in the database: %v\n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}

func (a *ActorModel) GetActorByIdWithMovies(db *sql.DB, uuid uuid.UUID) (ActorModel, error) {
	log.Printf("Getting actor with uuid %s in DB with movies... \n", uuid)

	query := `SELECT 
		a.id, a.name, a.surname, a.birthday, 
		a.created_at, a.updated_at, a.deleted_at, 
		a.creator_id, 
		m.id, m.title, m.director,
		m.release_date, m.average_grade
        	FROM actors a
        		LEFT JOIN movies m ON m.actor_id = a.id
        			WHERE a.id = $1; `

	var actor ActorModel
	rows, err := db.Query(query, uuid)
	if err != nil {
		log.Printf("Error getting actor by id with movies from database: %v\n", err)
		return ActorModel{}, err
	}
	defer rows.Close()

	movies := make([]MovieModel, 0)
	for rows.Next() {
		var movie MovieModel
		err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId, &movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade)
		if err != nil {
			log.Printf("Error scanning movie row in GetActorByIdWithMovies: %v\n", err)
			continue
		}
		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows in GetActorByIdWithMovies: %v\n", err)
		return ActorModel{}, err
	}

	actor.Movies = movies

	return actor, nil
}

func (a *ActorModel) DeleteActorById(db *sql.DB, uuid uuid.UUID) error {
	log.Printf("Deleting actor with uuid %s in DB... \n", uuid)

	query := `UPDATE actors 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at ISNULL;`

	_, err := db.Exec(query, uuid)
	if err != nil {
		log.Printf("Error deleting actor by uuid: %v\n", err)
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

	updateQueryBuilder.WriteString("updated_at = CURRENT_TIMESTAMP, ")
	query := strings.TrimSuffix(updateQueryBuilder.String(), ", ")
	query += " WHERE id = $" + strconv.Itoa(argIndex) + " AND deleted_at IS NULL RETURNING id, name, surname, birthday, created_at, updated_at, deleted_at, creator_id;"
	args = append(args, uuid)

	var actor ActorResponse
	err := db.QueryRow(query, args...).Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId)
	if err != nil {
		log.Printf("Error updating actor by uuid: %v\n", err)
		return ActorResponse{}, err
	}

	return actor, nil
}
