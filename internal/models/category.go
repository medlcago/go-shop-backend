package models

import (
	"database/sql"

	"github.com/google/uuid"
)

type ProductCategory struct {
	ID          uuid.UUID           `db:"id"`
	Name        string              `db:"name"`
	Slug        string              `db:"slug"`
	ParentID    sql.Null[uuid.UUID] `db:"parent_id"`
	HasChildren bool                `db:"has_children"`
}
