package initializers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
    // Setup
    if err := Setup(); err != nil {
        log.Fatalf("Error setting up tests: %v", err)
    }

    // Run tests
    exitCode := m.Run()

    // Teardown
    if err := Teardown(); err != nil {
        log.Fatalf("Error tearing down tests: %v", err)
    }

    os.Exit(exitCode)
}

func Setup() error {
	// Initializing env variables
	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal("Error initializing environment variables", err)
	}
    
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
        return fmt.Errorf("error connecting to PostgreSQL: %v", err)
    }
    defer db.Close()

    // Create test database
	testDb := "testdb"
    if _, err := db.Exec("CREATE DATABASE " + testDb + ";"); err != nil {
        return fmt.Errorf("error creating test database: %v", err)
    }

    return nil
}

func Teardown() error {
    // Read environment variables from .env file
	var (
		user     string = os.Getenv("PGUSER")
		password string = os.Getenv("PGPASSWORD")
		host     string = os.Getenv("PGHOST")
		port     string = os.Getenv("PGPORT")
		dbName   string = "testdb"
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

func Test_NewDatabaseConn(t *testing.T) {
	os.Setenv("PGDATABASE", "testdb")

	// Call the function being tested
	db := NewDatabaseConn()
	defer db.Close()

	// Assert that the connection is not nil
	if db == nil {
		t.Errorf("Expected a non-nil database connection, got nil")
	}

	if err := db.Ping(); err != nil {
		t.Errorf("Error pinging db to make sure it works: %v", err)
	}
}
