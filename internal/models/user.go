package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleSeller   UserRole = "seller"
	UserRoleCustomer UserRole = "customer"
)

type User struct {
	ID           uuid.UUID      `db:"id"`
	Email        string         `db:"email"`
	PasswordHash string         `db:"password_hash"`
	FullName     sql.NullString `db:"full_name"`
	Phone        sql.NullString `db:"phone"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
	Role         UserRole       `db:"role"`
}
