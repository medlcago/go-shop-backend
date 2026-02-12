package response

import "github.com/gofiber/fiber/v3"

type Response[T any] struct {
	Result  T      `json:"result"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

func NewResponse[T any](data T) *Response[T] {
	return &Response[T]{
		Result: data,
	}
}

func JSON[T any](ctx fiber.Ctx, status int, data T) error {
	return ctx.Status(status).JSON(NewResponse(data))
}
