package response

import "github.com/gofiber/fiber/v3"

type PaginatedResponse[T any] struct {
	Data  T     `json:"data"`
	Total int64 `json:"total"`
}

func NewPaginated[T any](data T, total int64) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Data:  data,
		Total: total,
	}
}

func PaginatedJSON[T any](ctx fiber.Ctx, status int, data T, total int64) error {
	return ctx.Status(status).JSON(NewResponse(NewPaginated(data, total)))
}
