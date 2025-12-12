package advice

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

func (h *Handler) GetAdvice(c echo.Context) error {
	var req AdviceRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}

	if req.Question == "" {
		return apperrors.NewWithDetails(400, "Пожалуйста, введите вопрос", "question field is required")
	}

	answer, err := h.service.GetAdvice(req.Question)
	if err != nil {
		return err
	}

	return c.JSON(200, AdviceResponse{
		Answer: answer,
	})
}

func (h *Handler) Analyze(c echo.Context) error {
	var req AnalysisRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.NewWithDetails(400, "Неверный формат запроса", err.Error())
	}

	// Валидация обязательных полей
	if req.Status == "" || req.Expenses == "" || req.Income == "" {
		return apperrors.NewWithDetails(400, "Пожалуйста, заполните все обязательные поля", "status, expenses, and income are required")
	}

	result, err := h.service.AnalyzeFinances(req)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

