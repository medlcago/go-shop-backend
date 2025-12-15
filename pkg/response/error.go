package response

import "github.com/gofiber/fiber/v3"

func NewError(err string) *Response[struct{}] {
	return &Response[struct{}]{
		Error: err,
	}
}

func Error(ctx fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(NewResponse(struct{}{}, message))
}
