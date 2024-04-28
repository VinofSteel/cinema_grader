package models

import (
	"database/sql"
	"log"
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

func (a *ActorModel) GetActorById(db *sql.DB, uuid uuid.UUID) (ActorResponse, error) {
	log.Printf("Getting user with uuid %s in DB... \n", uuid)

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
	query := `SELECT 
		a.id, a.name, a.surname, a.birthday, 
		a.created_at, a.updated_at, a.deleted_at, 
		a.creator_id, 
		m.id, m.title, m.director,
		m.release_date, m.grade
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
		err := rows.Scan(&actor.ID, &actor.Name, &actor.Surname, &actor.Birthday, &actor.CreatedAt, &actor.UpdatedAt, &actor.DeletedAt, &actor.CreatorId, &movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.Grade)
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
