package models

import (
	"database/sql"
	"log"
	"time"
)

type UserModel struct {
	ID        string       `json:"id"`
	Name      string       `json:"name" validate:"required"`
	Surname   string       `json:"surname" validate:"omitempty"`
	Email     string       `json:"email" validate:"required,email"`
	Password  string       `json:"password" validate:"required,password"`
	Birthday  string       `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
	IsAdm     bool         `json:"isAdm"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`
}

type UserBody struct {
	Name     string `json:"name" validate:"required"`
	Surname  string `json:"surname" validate:"omitempty"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
	Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Birthday  string    `json:"birthday"`
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

func (u *UserModel) GetUserByEmail(db *sql.DB, email string) (UserModel, error) {
	log.Printf("Getting user with emal %s in DB... \n", email)

	query := `SELECT id, name, surname, email, is_adm, created_at, updated_at, deleted_at FROM users WHERE email = $1`

	var user UserModel
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.IsAdm, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		log.Printf("Error getting user by email: %v\n", err)
		return UserModel{}, err
	}

	return user, nil
}
