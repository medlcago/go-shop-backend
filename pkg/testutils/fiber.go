package testutils

import (
	"encoding/json"

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

func CreateTestApp(cfg ...fiber.Config) *fiber.App {
	app := fiber.New(cfg...)
	return app
}
