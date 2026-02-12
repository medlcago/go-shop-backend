package response

import (
	"github.com/gofiber/fiber/v3"
)

func NewError(err string, details ...any) *Response[struct{}] {
	resp := &Response[struct{}]{
		Error: err,
	}

	if len(details) > 0 {
		resp.Details = details[0]
	}

	return resp
}

func Error(ctx fiber.Ctx, status int, err string, details ...any) error {
	return ctx.Status(status).JSON(NewError(err, details...))
}
