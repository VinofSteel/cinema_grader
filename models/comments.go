package models

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
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

var userModel UserModel
var movieModel MovieModel

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

func (c *CommentModel) GetAllComments(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]CommentResponse, error) {
	log.Printf("Getting all comments in DB, with offset %v, limit %v, orderBy %v and deleted %v...\n", offset, limit, orderBy, deleted)

	var getCommentsQueryBuilder strings.Builder
	getCommentsQueryBuilder.WriteString(`SELECT 
	id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id 
	FROM comments`)

	if !deleted {
		getCommentsQueryBuilder.WriteString(" WHERE deleted_at IS NULL")
	}

	getCommentsQueryBuilder.WriteString(" ORDER BY " + orderBy + " OFFSET $1 LIMIT $2;")

	query := getCommentsQueryBuilder.String()
	rows, err := db.Query(query, offset, limit)
	if err != nil {
		log.Println("Error getting all comments from db:", err)
		return nil, err
	}
	defer rows.Close()

	var comments []CommentResponse
	for rows.Next() {
		var comment CommentResponse
		if err := rows.Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
			log.Println("Error scanning comment from db:", err)
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (c *CommentModel) GetCommentById(db *sql.DB, uuid uuid.UUID) (CommentResponse, error) {
	log.Printf("Getting comment with uuid %s in DB... \n", uuid)

	query := `SELECT 
		id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id 
        FROM comments
        	WHERE id = $1;`

	var comment CommentResponse
	if err := db.QueryRow(query, uuid).Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
		log.Printf("Error getting comment by id in the database: %v\n", err)
		return CommentResponse{}, err
	}

	return comment, nil
}

func (c *CommentModel) DeleteCommentById(db *sql.DB, uuid uuid.UUID) error {
	log.Printf("Deleting comment with uuid %s in DB... \n", uuid)

	query := `UPDATE comments 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL;`

	_, err := db.Exec(query, uuid)
	if err != nil {
		log.Printf("Error deleting comment by uuid: %v\n", err)
		return err
	}

	return nil
}

func (c *CommentModel) UpdateCommentsById(db *sql.DB, uuid uuid.UUID, body CommentEditBody) (CommentResponse, error) {
	log.Printf("Updating comment with uuid %s in DB... \n", uuid)

	var updateQueryBuilder strings.Builder
	var args []interface{}

	updateQueryBuilder.WriteString("UPDATE comments SET ")

	argIndex := 1
	if body.Comment != "" {
		updateQueryBuilder.WriteString("comment = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Comment)
		argIndex++
	}

	if body.Grade != 0.0 {
		updateQueryBuilder.WriteString("grade = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Grade)
		argIndex++
	}

	updateQueryBuilder.WriteString("updated_at = CURRENT_TIMESTAMP, ")
	query := strings.TrimSuffix(updateQueryBuilder.String(), ", ")
	query += " WHERE id = $" + strconv.Itoa(argIndex) + " AND deleted_at IS NULL RETURNING id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id;"
	args = append(args, uuid)

	var comment CommentResponse
	if err := db.QueryRow(query, args...).Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
		log.Printf("Error updating comment by uuid: %v \n", err)
		return CommentResponse{}, err
	}

	return comment, nil
}

func (c *CommentModel) GetAllUserCommentsInDb(db *sql.DB, uuid uuid.UUID, orderBy string, deleted bool) (UserResponseWithComments, error) {
	user, err := userModel.GetUserById(db, uuid)
	if err != nil {
		log.Printf("Error getting user info of user %v from db: %v \n", uuid, err)
		return UserResponseWithComments{}, err
	}

	var getCommentsQueryBuilder strings.Builder
	getCommentsQueryBuilder.WriteString(`SELECT 
		id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id
		FROM comments
			WHERE user_id = $1`)

	if !deleted {
		getCommentsQueryBuilder.WriteString(" AND deleted_at IS NULL")
	}

	getCommentsQueryBuilder.WriteString(" ORDER BY " + orderBy + ";")

	query := getCommentsQueryBuilder.String()
	rows, err := db.Query(query, uuid)
	if err != nil {
		log.Printf("Error getting all comments of user %v from db: %v \n", uuid, err)
		return UserResponseWithComments{}, err
	}
	defer rows.Close()

	userWithComments := UserResponseWithComments{
		UserResponse: user,
		Comments:     []CommentResponse{},
	}
	for rows.Next() {
		var comment CommentResponse
		if err := rows.Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
			log.Printf("Error scanning rows while getting all comments of user %v from db: %v \n", uuid, err)
			return UserResponseWithComments{}, err
		}
		userWithComments.Comments = append(userWithComments.Comments, comment)
	}

	return userWithComments, nil
}

func (c *CommentModel) GetAllCommentsInAMovieInDb(db *sql.DB, uuid uuid.UUID, orderBy string, deleted bool) (MovieResponseWithActorsWithComments, error) {
	movie, err := movieModel.GetMovieByIdWithActors(db, uuid)
	if err != nil {
		log.Printf("Error getting movie info of movie %v from db: %v \n", uuid, err)
		return MovieResponseWithActorsWithComments{}, err
	}

	var getCommentsQueryBuilder strings.Builder
	getCommentsQueryBuilder.WriteString(`SELECT 
		id, comment, grade, created_at, updated_at, deleted_at, user_id, movie_id
		FROM comments
			WHERE movie_id = $1`)

	if !deleted {
		getCommentsQueryBuilder.WriteString(" AND deleted_at IS NULL")
	}

	getCommentsQueryBuilder.WriteString(" ORDER BY " + orderBy + ";")

	query := getCommentsQueryBuilder.String()
	rows, err := db.Query(query, uuid)
	if err != nil {
		log.Printf("Error getting all comments of user %v from db: %v \n", uuid, err)
		return MovieResponseWithActorsWithComments{}, err
	}
	defer rows.Close()

	movieWithComments := MovieResponseWithActorsWithComments{
		MovieResponseWithActors: movie,
		Comments:                []CommentResponse{},
	}
	for rows.Next() {
		var comment CommentResponse
		if err := rows.Scan(&comment.ID, &comment.Comment, &comment.Grade, &comment.CreatedAt, &comment.UpdatedAt, &comment.DeletedAt, &comment.UserId, &comment.MovieId); err != nil {
			log.Printf("Error scanning rows while getting all comments of user %v from db: %v \n", uuid, err)
			return MovieResponseWithActorsWithComments{}, err
		}
		movieWithComments.Comments = append(movieWithComments.Comments, comment)
	}

	return movieWithComments, nil
}
