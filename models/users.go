package models

import (
	"database/sql"
	"log"
	"time"
)

type UserModel struct {
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Birthday  time.Time `json:"birthday"`
	IsAdm     bool      `json:"isAdm"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

type UserBody struct {
	Name     string    `json:"name"`
	Surname  string    `json:"surname"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Birthday time.Time `json:"birthday"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Birthday  time.Time `json:"birthday"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (u *UserModel) InsertUserInDB(db *sql.DB, userInfo UserBody) (UserResponse, error) {
	log.Printf("Inserting user with email %s in DB...\n", userInfo.Email)

	query := `INSERT INTO users(name, surname, email, password, birthday)
              VALUES ($1, $2, $3, $4, $5) RETURNING id, name, surname, email, birthday, created_at, updated_at;`

	var user UserResponse
	err := db.QueryRow(query, userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.Password, userInfo.Birthday).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Birthday, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		return UserResponse{}, err
	}

	return user, nil
}
