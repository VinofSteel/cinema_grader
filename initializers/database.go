package initializers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitializeDB() *sql.DB  {
	var (
		user = os.Getenv("PGUSER")
		password = os.Getenv("PGPASSWORD")
		host = os.Getenv("PGHOST")
		port = os.Getenv("PGPORT")
		dbName = os.Getenv("PGDATABASE")
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