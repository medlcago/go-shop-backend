package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID          uuid.UUID           `db:"id"`
	Name        string              `db:"name"`
	Slug        string              `db:"slug"`
	ParentID    sql.Null[uuid.UUID] `db:"parent_id"`
	HasChildren bool                `db:"has_children"`
}

type Product struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Price       float64        `db:"price"`
	Slug        string         `db:"slug"`
	Stock       int            `db:"stock"`
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
