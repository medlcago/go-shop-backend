package response

import "github.com/gofiber/fiber/v3"

type Response[T any] struct {
	Result T      `json:"result"`
	Error  string `json:"error"`
}

func NewResponse[T any](data T, err string) *Response[T] {
	return &Response[T]{
		Result: data,
		Error:  err,
	}
}

func JSON[T any](ctx fiber.Ctx, status int, data T) error {
	return ctx.Status(status).JSON(NewResponse(data, ""))
}
