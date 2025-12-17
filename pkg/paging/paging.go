package paging

const (
	DefaultPageSize = 50
)

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func New(limit, offset int) *Pagination {
	if limit <= 0 || limit > DefaultPageSize {
		limit = DefaultPageSize
	}

	if offset < 0 {
		offset = 0
	}

	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}
