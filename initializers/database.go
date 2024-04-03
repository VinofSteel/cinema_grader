package initializers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	log.Printf("Opening connection with database %s on port %s,,.\n", dbName, port)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Error opening db connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging db: %v", err)
	}

	log.Println("Connection opened succesfully!")

	return db
}

func CreateTables(db *sql.DB) {
	// Creating uuid extension on DB
	extensionQuery := "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
	_, err := db.Exec(extensionQuery)
	if err != nil {
		log.Fatal(err)
	}
}
