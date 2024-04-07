package initializers

import (
	"log"
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/tests"
)

func TestMain(m *testing.M) {
    // Setup
    if err := tests.Setup(); err != nil {
        log.Fatalf("Error setting up tests: %v", err)
    }

    // Run tests
    exitCode := m.Run()

    // Teardown
    if err := tests.Teardown(); err != nil {
        log.Fatalf("Error tearing down tests: %v", err)
    }

    os.Exit(exitCode)
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
