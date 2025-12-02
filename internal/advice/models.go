package advice

type AdviceRequest struct {
	Question string `json:"question" validate:"required"`
}

type AdviceResponse struct {
	Answer string `json:"answer"`
}

// Новые модели для финансового анализа
type AnalysisRequest struct {
	Status     string  `json:"status" validate:"required"`
	Expenses   string  `json:"expenses" validate:"required"`
	Income     string  `json:"income" validate:"required"`
	Additional *string `json:"additional"`
}

type AnalysisResponse struct {
	Balance string `json:"balance"`
	Advice  string `json:"advice"`
}

