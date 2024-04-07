package initializers

import (
	"log"
	"os"
	"testing"

	"github.com/VinOfSteel/cinemagrader/tests"
	"github.com/stretchr/testify/assert"
)

var testDb string

func TestMain(m *testing.M) {
	// Setup
	var err error
	testDb, err = tests.Setup()
	if err != nil {
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
	os.Setenv("PGDATABASE", testDb)

	// Call the function being tested
	db := NewDatabaseConn()
	defer db.Close()

	// Assert that the connection is not nil
	assert.NotNil(t, db, "Expected a non-nil database connection")

	err := db.Ping()
	assert.NoError(t, err, "Error pinging db to make sure it works: %v", err)
}
