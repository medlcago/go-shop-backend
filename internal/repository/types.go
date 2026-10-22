package repository

import "github.com/google/uuid"

type LimitOffsetFilter struct {
	Limit  int
	Offset int
}

type GetUserFilter struct {
	IncludeDeleted bool
}

type ListProductFilter struct {
	LimitOffsetFilter
	OrderBy    string
	OrderDesc  bool
	CategoryID uuid.UUID
	MinPrice   int64
	MaxPrice   int64
}

type SearchProductFilter struct {
	LimitOffsetFilter
	Query string
}

type ListCategoryFilter struct {
	LimitOffsetFilter
	ID uuid.UUID
}

type ListOrderFilter struct {
	LimitOffsetFilter
	Status string
}

type ListWishlistFilter struct {
	LimitOffsetFilter
}
