package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type MovieModel struct {
	ID           uuid.UUID    `json:"id"`
	Title        string       `json:"title"`
	Director     string       `json:"director"`
	ReleaseDate  string       `json:"releaseDate"`
	AverageGrade float64      `json:"averageGrade"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DeletedAt    sql.NullTime `json:"deletedAt"`

	CreatorId string       `json:"creatorId"`
	Actors    []ActorModel `json:"actors"`
	Comments  []CommentModel `json:"comments"`
}

type MovieBody struct {
	Title        string       `json:"title" validate:"required"`
	Director     string       `json:"director" validate:"required"`
	ReleaseDate  string       `json:"releaseDate" validate:"omitempty,datetime=2006-01-02"`

	CreatorId string       	  `json:"creatorId" validate:"required,isAdminUuid"`
}

type MovieResponse struct {
	ID           uuid.UUID    `json:"id"`
	Title        string       `json:"title"`
	Director     string       `json:"director"`
	ReleaseDate  string       `json:"releaseDate"`
	AverageGrade float64      `json:"averageGrade"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	DeletedAt    sql.NullTime `json:"deletedAt"`

	CreatorId string       	  `json:"creatorId"`
}

func (m *MovieModel) InsertMovieInDB(db *sql.DB, movieInfo MovieBody) (MovieResponse, error) {
	log.Printf("Inserting movie with title %s in DB by user %s...\n", movieInfo.Director, movieInfo.CreatorId)

	query := `INSERT INTO actors
			(title, director, release_date, creator_id)
			VALUES ($1, $2, $3, $4)
				RETURNING id, title, director, release_date, average_grade, created_at, updated_at, deleted_at, creator_id;`

	var movie MovieResponse
	err := db.QueryRow(query, movieInfo.Title, movieInfo.Director, movieInfo.ReleaseDate, movieInfo.CreatorId).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
	if err != nil {
		log.Printf("Error inserting movie into database: %v\n", err)
		return MovieResponse{}, err
	}

	return movie, nil
}
