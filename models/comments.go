package models

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type CommentModel struct {
	ID        uuid.UUID    `json:"id"`
	Comment   string       `json:"comment"`
	Grade     float64      `json:"grade"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	UserId  string `json:"userId"`
	MovieId string `json:"movieId"`
}

type CommentBody struct {
	Comment string  `json:"comment" validate:"required"`
	Grade   float64 `json:"grade" validate:"omitempty,isvalidgrade"`

	MovieId string `json:"movieId" validate:"required,isvaliduuid"`
}

type CommentEditBody struct {
	Comment string  `json:"comment" validate:"omitempty"`
	Grade   float64 `json:"grade" validate:"omitempty"`
}

type CommentResponse struct {
	ID        uuid.UUID    `json:"id"`
	Comment   string       `json:"comment"`
	Grade     float64      `json:"grade"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	UserId  string `json:"userId"`
	MovieId string `json:"movieId"`
}

// Public methods
func (c *CommentModel) InsertCommentInDB(db *sql.DB, uuid uuid.UUID, commentInfo CommentBody) (CommentResponse, error) {
	log.Printf("Inserting comment in DB by user %s...\n", uuid)

	query := `INSERT INTO comments
			(comment, grade, user_id, movie_id)
			VALUES ($1, $2, $3, $4)
				RETURNING id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id;`

	var comment CommentResponse
	if err := db.QueryRow(query, commentInfo.Comment, commentInfo.Grade, uuid, commentInfo.MovieId).Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
		log.Printf("Error inserting comment into database: %v\n", err)
		return CommentResponse{}, err
	}

	return comment, nil
}
