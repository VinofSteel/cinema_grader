package tests

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/VinOfSteel/cinemagrader/models"
	"github.com/joho/godotenv"
)

var testDb string
var UserModel models.UserModel

func Setup() (string, error) {
	// Initializing env variables
	func() {
		err := godotenv.Load("../.env")

		if err != nil {
			log.Fatal("Error initializing environment variables", err)
		}
	}()

	// Generating a random string to be the test database name.
	// This is done because all tests run in paralel, meaning that we would be creating
	// a bunch of DBs with the same name.
	func() {
		const letterBytes = "abcdefghijklmnopqrstuvwxyz"

		b := []byte{'t', 'e', 's', 't', '_'}
		for len(b) < 15 {
			b = append(b, letterBytes[rand.Intn(len(letterBytes))])
		}

		testDb = string(b)
	}()

	var (
		user     string = os.Getenv("PGUSER")
		password string = os.Getenv("PGPASSWORD")
		host     string = os.Getenv("PGHOST")
		port     string = os.Getenv("PGPORT")
		dbName   string = os.Getenv("PGDATABASE")
	)

	// Connect to PostgreSQL
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return "", fmt.Errorf("error connecting to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Create test database
	if _, err := db.Exec("CREATE DATABASE " + testDb + ";"); err != nil {
		return "", fmt.Errorf("error creating test database: %v", err)
	}

	return testDb, nil
}

func Teardown() error {
	// Read environment variables from .env file
	var (
		user     string = os.Getenv("PGUSER")
		password string = os.Getenv("PGPASSWORD")
		host     string = os.Getenv("PGHOST")
		port     string = os.Getenv("PGPORT")
		dbName   string = testDb
	)

	// Connect to PostgreSQL
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, "postgres")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Close the connection to the test database
	if _, err := db.Exec("SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = $1 AND pid <> pg_backend_pid();", dbName); err != nil {
		return fmt.Errorf("error terminating connections to test database: %v", err)
	}

	// Drop test database
	if _, err := db.Exec("DROP DATABASE IF EXISTS " + dbName + ";"); err != nil {
		return fmt.Errorf("error dropping test database: %v", err)
	}

	return nil
}

// God, forgive me for what I'm about to do
func InsertMockedUsersInDb(db *sql.DB, users []models.UserBody, respChan *chan models.UserResponse) {
	var wg sync.WaitGroup
	
	for _, user := range users {
		wg.Add(1)
		go func(user models.UserBody) {
			defer wg.Done()
			userResp, err := UserModel.InsertUserInDB(db, user)
			if err != nil {
				log.Fatalf("Error inserting mocked user with email %v in Db: %v", user.Email, err)
			}
			*respChan <- userResp
		}(user)
	}

	wg.Wait()
	close(*respChan)
	// You know when you do something to not have to do another thing to save time and you end up wasting more time than you would have if you just did the original thing?
	// Yeah.
}
