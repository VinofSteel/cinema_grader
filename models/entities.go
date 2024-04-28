package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const UsersTable string = `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(50) NOT NULL,
		surname VARCHAR(70),
		email VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(200) NOT NULL,
		birthday DATE,
		is_adm BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
`

type MovieModel struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Director    string       `json:"director"`
	ReleaseDate time.Time    `json:"releaseData"`
	Grade       float64      `json:"grade"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
	DeletedAt   sql.NullTime `json:"deletedAt"`

	CreatorId string       `json:"creatorId"`
	Actors    []ActorModel `json:"actors"`
}

const MoviesTable string = `
	CREATE TABLE IF NOT EXISTS movies (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		title VARCHAR(50) NOT NULL,
		director VARCHAR(50) NOT NULL,
		release_date DATE NOT NULL,
		grade DECIMAL(3, 1),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,
		
		creator_id UUID NOT NULL,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);
`

const ActorsTable string = `
	CREATE TABLE IF NOT EXISTS actors (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(50) NOT NULL,
		surname VARCHAR(70),
		birthday DATE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,

		creator_id UUID NOT NULL,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);
`

const MoviesActorsPivotTable string = `
	CREATE TABLE IF NOT EXISTS movies_actors(
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		
		actor_id UUID NOT NULL,
		movie_id UUID,
		FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE RESTRICT,
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE RESTRICT
	);
`

type Comment struct {
	ID        uuid.UUID    `json:"id"`
	Comment   string       `json:"comment"`
	Grade     float64      `json:"grade"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`

	UserId  string `json:"userId"`
	MovieId string `json:"movieId"`
}

const CommentsTable string = `
	CREATE TABLE IF NOT EXISTS comments(
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		comment TEXT NOT NULL,
		grade DECIMAL(3, 1),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,
		
		user_id UUID NOT NULL,
		movie_id UUID NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE RESTRICT
	);
`

var Tables = []string{UsersTable, MoviesTable, ActorsTable, MoviesActorsPivotTable, CommentsTable}
