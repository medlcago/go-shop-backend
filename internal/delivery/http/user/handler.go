package user

import (
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/middleware"
	"go-shop-backend/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	userService service.UserService
}

func NewHandler(userService service.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// Login godoc
//
//	@Summary		Login
//	@Description	Login
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserLoginRequest	true	"Request body for login"
//	@Success		200		{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/users/login [post]
func (h *Handler) Login(ctx fiber.Ctx) error {
	var req dto.UserLoginRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.userService.Login(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Register godoc
//
//	@Summary		Register
//	@Description	Register
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.UserRegisterRequest	true	"Request body for registration"
//	@Success		201		{object}	response.Response[dto.UserTokenResponse]
//	@Failure		400		{object}	response.Response[any]
//	@Failure		409		{object}	response.Response[any]	"The user already exists"
//	@Failure		500		{object}	response.Response[any]
//	@Router			/users/register [post]
func (h *Handler) Register(ctx fiber.Ctx) error {
	var req dto.UserRegisterRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.userService.Register(ctx, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusCreated, resp)
}

// Setup2FA godoc
//
//	@Summary		Setup 2FA
//	@Description	Initialize two-factor authentication setup for the authenticated user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[dto.Setup2FAResponse]
//	@Failure		400	{object}	response.Response[any]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		409	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/users/me/setup-2fa [post]
func (h *Handler) Setup2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.Setup2FA(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

// Confirm2FA godoc
//
//	@Summary		Confirm 2FA
//	@Description	Confirm and enable two-factor authentication with the provided code
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Confirm2FARequest		true	"Request body with 2FA confirmation code"
//	@Success		200		{object}	response.Response[string]	"2FA successfully enabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/users/me/confirm-2fa [post]
func (h *Handler) Confirm2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Confirm2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.userService.Confirm2FA(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// Disable2FA godoc
//
//	@Summary		Disable 2FA
//	@Description	Disable two-factor authentication for the authenticated user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Disable2FARequest		true	"Request body with password confirmation"
//	@Success		200		{object}	response.Response[string]	"2FA successfully disabled"
//	@Failure		400		{object}	response.Response[any]
//	@Failure		401		{object}	response.Response[any]
//	@Failure		500		{object}	response.Response[any]
//	@Router			/users/me/disable-2fa [post]
func (h *Handler) Disable2FA(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.Disable2FARequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	err := h.userService.Disable2FA(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, "OK")
}

// GetMe godoc
//
//	@Summary		Get Me
//	@Description	Get Me
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response[internal_dto.UserResponse]
//	@Failure		401	{object}	response.Response[any]
//	@Failure		404	{object}	response.Response[any]
//	@Failure		500	{object}	response.Response[any]
//	@Router			/users/me [get]
func (h *Handler) GetMe(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.GetUserByID(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) SendEmailConfirmationCode(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	resp, err := h.userService.SendEmailConfirmationCode(ctx, *userCtx.UserID)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}

func (h *Handler) ConfirmEmail(ctx fiber.Ctx) error {
	userCtx := middleware.GetUserContext(ctx)
	if userCtx.UserID == nil {
		return apperror.ErrInvalidCredentials
	}

	var req dto.ConfirmEmailRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return err
	}

	resp, err := h.userService.ConfirmEmail(ctx, *userCtx.UserID, req)
	if err != nil {
		return err
	}

	return response.JSON(ctx, fiber.StatusOK, resp)
}
