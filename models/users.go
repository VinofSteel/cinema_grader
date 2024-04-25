package models

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	ID        uuid.UUID    `json:"id"`
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

type UserEditBody struct {
	Name     string `json:"name" validate:"omitempty"`
	Surname  string `json:"surname" validate:"omitempty"`
	Password string `json:"password" validate:"omitempty,password"`
	Birthday string `json:"birthday" validate:"omitempty,datetime=2006-01-02"`
}

type UserResponse struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Surname   string       `json:"surname"`
	Email     string       `json:"email"`
	Birthday  string       `json:"birthday"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	DeletedAt sql.NullTime `json:"deletedAt"`
}

func (u *UserModel) InsertUserInDB(db *sql.DB, userInfo UserBody) (UserResponse, error) {
	log.Printf("Inserting user with email %s in DB...\n", userInfo.Email)

	query := `INSERT INTO users
			(name, surname, email, password, birthday)
            VALUES ($1, $2, $3, $4, $5) 
			  	RETURNING id, name, surname, email, birthday, created_at, updated_at;`

	var user UserResponse
	err := db.QueryRow(query, userInfo.Name, userInfo.Surname, userInfo.Email, userInfo.Password, userInfo.Birthday).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Birthday, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		return UserResponse{}, err
	}

	return user, nil
}

func (u *UserModel) GetUserByEmail(db *sql.DB, email string) (UserModel, error) {
	log.Printf("Getting user with email %s in DB... \n", email)

	query := `SELECT 
		id, name, surname, email, password, is_adm, created_at, updated_at, deleted_at 
		FROM users 
			WHERE email = $1;`

	var user UserModel
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Password, &user.IsAdm, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		log.Printf("Error getting user by email: %v\n", err)
		return UserModel{}, err
	}

	return user, nil
}

func (u *UserModel) GetAllUsers(db *sql.DB, offset, limit int, orderBy string, deleted bool) ([]UserResponse, error) {
	log.Printf("Getting all users in DB, with offset %v, limit %v and orderBy %v...", offset, limit, orderBy)

	query := `SELECT 
		id, name, surname, email, birthday, created_at, updated_at, deleted_at 
		FROM users`

	if !deleted {
		query += ` WHERE deleted_at ISNULL`
	}

	query += ` ORDER BY ` + orderBy + ` OFFSET $1 LIMIT $2;`

	rows, err := db.Query(query, offset, limit)
	if err != nil {
		log.Println("Error getting all users from db:", err)
		return nil, err
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		if err := rows.Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Birthday, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u *UserModel) GetUserById(db *sql.DB, uuid uuid.UUID) (UserResponse, error) {
	log.Printf("Getting user with uuid %s in DB... \n", uuid)

	query := `SELECT 
		id, name, surname, email, birthday, created_at, updated_at 
		FROM users 
			WHERE id = $1;`

	var user UserResponse
	err := db.QueryRow(query, uuid).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Birthday, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("Error getting user by uuid: %v\n", err)
		return UserResponse{}, err
	}

	return user, nil
}

func (u *UserModel) DeleteUserById(db *sql.DB, uuid uuid.UUID) error {
	log.Printf("Deleting user with uuid %s in DB... \n", uuid)

	query := `UPDATE users 
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at ISNULL;`

	_, err := db.Exec(query, uuid)
	if err != nil {
		log.Printf("Error deleting user by uuid: %v\n", err)
		return err
	}

	return nil
}

func (u *UserModel) UpdateUserById(db *sql.DB, uuid uuid.UUID, body UserEditBody) (UserResponse, error) {
	log.Printf("Updating user with uuid %s in DB... \n", uuid)

	var updateQuery strings.Builder
	var args []interface{}

	updateQuery.WriteString("UPDATE users SET ")

	argIndex := 1
	if body.Name != "" {
		updateQuery.WriteString("name = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Name)
		argIndex++
	}

	if body.Surname != "" {
		updateQuery.WriteString("surname = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Surname)
		argIndex++
	}

	if body.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)

		if err != nil {
			log.Println("Error encrypting user's password while updating it:", err)
			return UserResponse{}, err
		}

		updateQuery.WriteString("password = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, string(hashedPassword))
		argIndex++
	}

	if body.Birthday != "" {
		updateQuery.WriteString("birthday = $" + strconv.Itoa(argIndex) + ", ")
		args = append(args, body.Birthday)
		argIndex++
	}

	updateQuery.WriteString("updated_at = CURRENT_TIMESTAMP, ")

	updateQueryString := strings.TrimSuffix(updateQuery.String(), ", ")
	updateQueryString += " WHERE id = $" + strconv.Itoa(argIndex) + " AND deleted_at IS NULL RETURNING id, name, surname, email, birthday, created_at, updated_at;"
	args = append(args, uuid)

	var user UserResponse
	err := db.QueryRow(updateQueryString, args...).Scan(&user.ID, &user.Name, &user.Surname, &user.Email, &user.Birthday, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("Error updating user by uuid: %v\n", err)
		return UserResponse{}, err
	}

	return user, nil
}
