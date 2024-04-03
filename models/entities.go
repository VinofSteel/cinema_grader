package models

import "time"

type User struct {
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Birthday  time.Time `json:"birthday"`
	IsAdm     bool      `json:"isAdm"`
	CratedAt  time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

const UserTable string = `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(50) NOT NULL,
		surname VARCHAR(70),
		email VARCHAR(100) NOT NULL,
		password VARCHAR(20) NOT NULL,
		birthday TIMESTAMP,
		is_adm BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		deleted_at TIMESTAMP
	);
`

var Tables = []string{UserTable}