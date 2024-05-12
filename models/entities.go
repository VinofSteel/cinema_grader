package models

const UsersTableQuery string = `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(50) NOT NULL,
		surname VARCHAR(70),
		email VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(200) NOT NULL,
		birthday DATE,
		is_adm BOOLEAN DEFAULT false,
		picture TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
`

const MoviesTableQuery string = `
	CREATE TABLE IF NOT EXISTS movies (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		title VARCHAR(50) NOT NULL UNIQUE,
		director VARCHAR(50) NOT NULL,
		release_date DATE NOT NULL,
		picture TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,
		
		creator_id UUID NOT NULL,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);
`

const ActorsTableQuery string = `
	CREATE TABLE IF NOT EXISTS actors (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(50) NOT NULL,
		surname VARCHAR(70),
		birthday DATE,
		picture TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,

		creator_id UUID NOT NULL,
		FOREIGN KEY (creator_id) REFERENCES users(id)
	);
`

const MoviesActorsPivotTableQuery string = `
	CREATE TABLE IF NOT EXISTS movies_actors(
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		
		actor_id UUID NOT NULL,
		movie_id UUID,
		FOREIGN KEY (actor_id) REFERENCES actors(id) ON DELETE RESTRICT,
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE RESTRICT
	);
`

const CommentsTableQuery string = `
	CREATE TABLE IF NOT EXISTS comments(
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		comment TEXT NOT NULL,
		grade DECIMAL(3, 1) CHECK (grade >= 1 AND grade <= 5),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP,
		
		user_id UUID NOT NULL,
		movie_id UUID NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
		FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE RESTRICT
	);
`

// Movies - comments grade relationship queries
// We need to create the average_grade column in the movies table that has the average grade of all comments made on that movie
// Since the comments table has to obligatorily be created after the movies table because of the relationship, we need to alter the movies table.

// 1 - Adding average_grade column
const MoviesAverageColumnQuery string = `
	ALTER TABLE movies ADD COLUMN IF NOT EXISTS average_grade DECIMAL(3, 1) DEFAULT 0;
`

// 2 - Creating SQL function to update average_grade column
const UpdateAverageGradeFunctionQuery string = `
	CREATE OR REPLACE FUNCTION update_average_grade()
	RETURNS TRIGGER AS $$
	BEGIN
		UPDATE movies
		SET average_grade = (
			SELECT COALESCE(AVG(grade), 0) FROM comments WHERE movie_id = NEW.movie_id
		)
		WHERE id = NEW.movie_id;

		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;
`

// 3 - Creating triggers to execute the function when inserting, updating or deleting data instead of doing it via code
const CommentInsertTriggerQuery string = `
	CREATE OR REPLACE TRIGGER comment_insert_trigger
	AFTER INSERT ON comments
		FOR EACH ROW EXECUTE FUNCTION update_average_grade();
`
const CommentUpdateTriggerQuery string = `
	CREATE OR REPLACE TRIGGER comment_update_trigger
	AFTER UPDATE ON comments
		FOR EACH ROW EXECUTE FUNCTION update_average_grade();
`

const CommentDeleteTriggerQuery string = `
	CREATE OR REPLACE TRIGGER comment_delete_trigger
	AFTER DELETE ON comments
		FOR EACH ROW EXECUTE FUNCTION update_average_grade();
`

var Queries = []string{UsersTableQuery, MoviesTableQuery, ActorsTableQuery, MoviesActorsPivotTableQuery, CommentsTableQuery, MoviesAverageColumnQuery, UpdateAverageGradeFunctionQuery, CommentInsertTriggerQuery, CommentUpdateTriggerQuery, CommentDeleteTriggerQuery}
