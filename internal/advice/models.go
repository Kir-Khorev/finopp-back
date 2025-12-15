package advice

type AdviceRequest struct {
	Question string `json:"question" validate:"required"`
}

type AdviceResponse struct {
	Answer string `json:"answer"`
}

type StructuredAdviceResponse struct {
	Answer            string  `json:"answer"`
	TotalIncomeRUB    float64 `json:"totalIncomeRUB"`
	TotalExpensesRUB  float64 `json:"totalExpensesRUB"`
	BalanceRUB        float64 `json:"balanceRUB"`
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

// Структурированные модели для конвертации валют
type FinanceSource struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type StructuredAdviceRequest struct {
	IncomeSources   []FinanceSource `json:"incomeSources" validate:"required,min=1"`
	ExpenseSources  []FinanceSource `json:"expenseSources" validate:"required,min=1"`
	Problems        []string        `json:"problems"`
	CustomProblem   string          `json:"customProblem"`
	AdditionalInfo  string          `json:"additionalInfo"`
}

