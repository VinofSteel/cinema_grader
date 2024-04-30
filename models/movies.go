package models

import (
	"database/sql"
	"log"
	"sort"
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

	CreatorId string      `json:"creatorId" validate:"required,isAdminUuid"`
	Actors    []uuid.UUID `json:"actors" validate:"required"`
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

	CreatorId string          `json:"creatorId"`
	Actors    []ActorResponse `json:"actors"`
}

var actorModel ActorModel

func (m *MovieModel) InsertMovieInDB(db *sql.DB, movieInfo MovieBody) (MovieResponse, error) {
	log.Printf("Inserting movie with title %s in DB by user %s...\n", movieInfo.Title, movieInfo.CreatorId)

	// Starting a transaction that can be rolled back if shit happens
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction to insert movie in DB: %v\n", err)
		return MovieResponse{}, err
	}
	defer tx.Rollback() // This will execute if any error is returned and cancel any changes to the db

	query := `INSERT INTO movies
			(title, director, release_date, creator_id)
			VALUES ($1, $2, $3, $4)
				RETURNING id, title, director, release_date, average_grade, created_at, updated_at, deleted_at, creator_id;`

	var movie MovieResponse
	err = tx.QueryRow(query, movieInfo.Title, movieInfo.Director, movieInfo.ReleaseDate, movieInfo.CreatorId).Scan(&movie.ID, &movie.Title, &movie.Director, &movie.ReleaseDate, &movie.AverageGrade, &movie.CreatedAt, &movie.UpdatedAt, &movie.DeletedAt, &movie.CreatorId)
	if err != nil {
		log.Printf("Error inserting movie into database: %v\n", err)
		return MovieResponse{}, err
	}

	log.Println("ACTORS IN MOVIE BODY:", movieInfo.Actors)

	// Associate actors with the movie in the pivot table
	var wg sync.WaitGroup
	errCh := make(chan error, len(movieInfo.Actors))
	actorInfoCh := make(chan ActorResponse, len(movieInfo.Actors))
	actorIndexMap := make(map[uuid.UUID]int)

	for i, actorID := range movieInfo.Actors {
		wg.Add(1)
		actorIndexMap[actorID] = i
		go func(actorID uuid.UUID) {
			defer wg.Done()

			actorResponse, err := actorModel.GetActorById(db, actorID)
			if err != nil {
				log.Printf("Trying to existing non-existant actor %v to a movie: %v\n", actorID, err)
				errCh <- err
				return
			}
			actorInfoCh <- actorResponse

			query := `INSERT INTO movies_actors (actor_id, movie_id) VALUES ($1, $2)`
			_, err = tx.Exec(query, actorID, movie.ID)
			if err != nil {
				log.Printf("Error associating actor %v with movie: %v\n", actorID, err)
				errCh <- err
				return
			}
		}(actorID)
	}

	wg.Wait()
	close(errCh)
	close(actorInfoCh)

	for err := range errCh {
		if err != nil {
			tx.Rollback()
			return MovieResponse{}, err
		}
	}

	// Getting the actors of the movie to fill the Actors slice
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
		return MovieResponse{}, err
	}

	return movie, nil
}
