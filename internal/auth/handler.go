package auth

import (
	apperrors "github.com/Kir-Khorev/finopp-back/pkg/errors"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}

	resp, err := h.service.Register(req)
	if err != nil {
		return err
	}

	return c.JSON(201, resp)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}

	resp, err := h.service.Login(req)
	if err != nil {
		return err
	}

	return c.JSON(200, resp)
}

