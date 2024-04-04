package initializers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/VinOfSteel/cinemagrader/models"
	_ "github.com/lib/pq"
)

func InitializeDB() *sql.DB {
	var (
		user     string = os.Getenv("PGUSER")
		password string = os.Getenv("PGPASSWORD")
		host     string = os.Getenv("PGHOST")
		port     string = os.Getenv("PGPORT")
		dbName   string = os.Getenv("PGDATABASE")
	)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)

	log.Printf("Opening connection with database %s on port %s...\n", dbName, port)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error opening db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging db: %v", err)
	}

	log.Println("Connection opened succesfully!")
	
	// Executing table creation queries as soon as DB is opened
	createTables(db)

	return db
}

func createTables(db *sql.DB) {
	// Creating uuid extension on DB
	extensionQuery := "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
	if _, err := db.Exec(extensionQuery); err != nil {
		log.Fatalf("Error creating uuid extension: %v", err)
	}

	// Creating application tables
	for i, query := range models.Tables {
		log.Printf("\"Creating\" table on index %v", i)
		if _, err := db.Exec(query); err != nil {
			log.Fatalf("Error creating tables: %v", err)
		}
	}
}
