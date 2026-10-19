package domain

import "time"

type Role string

const (
	RoleStudent    Role = "student"
	RoleInstructor Role = "instructor"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID           string
	Email        string
	Phone        string
	PasswordHash string
	FullName     string
	Role         Role
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(u *User) error
	GetByEmail(email string) (*User, error)
	GetByID(id string) (*User, error)
}
