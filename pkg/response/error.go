package response

import "github.com/gofiber/fiber/v3"

func Error(ctx fiber.Ctx, status int, message string) error {
	return ctx.Status(status).JSON(NewResponse(struct{}{}, message))
}
