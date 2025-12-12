package testutils

import (
	"encoding/json"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/validator"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
)

func PrepareTestContext(app *fiber.App, path string, body any) (fiber.Ctx, func()) {
	requestBody, _ := json.Marshal(body)
	var req fasthttp.Request
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(path)
	req.SetBody(requestBody)

	fctx := &fasthttp.RequestCtx{}
	fctx.Init(&req, nil, nil)

	ctx := app.AcquireCtx(fctx)
	ctx.Response().Reset()

	return ctx, func() {
		app.ReleaseCtx(ctx)
	}
}

func CreateTestApp() *fiber.App {
	l := logger.NewSlog(logger.EnvProduction)
	slog.SetDefault(l)

	app := fiber.New(fiber.Config{
		ErrorHandler:    middleware.ErrorHandler(),
		StructValidator: validator.New(),
	})

	return app
}
