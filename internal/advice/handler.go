package advice

import (
	"net/http"

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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Question == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Пожалуйста, введите вопрос.",
		})
	}

	answer, err := h.service.GetAdvice(req.Question)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, AdviceResponse{
		Answer: answer,
	})
}

func (h *Handler) Analyze(c echo.Context) error {
	var req AnalysisRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат запроса",
		})
	}

	// Валидация обязательных полей
	if req.Status == "" || req.Expenses == "" || req.Income == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Пожалуйста, заполните все обязательные поля (статус, расходы, доходы)",
		})
	}

	result, err := h.service.AnalyzeFinances(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}

