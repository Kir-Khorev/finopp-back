package advice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/Kir-Khorev/finopp-back/pkg/errors"
)

type CurrencyConverter interface {
	ConvertToRUB(ctx context.Context, amount float64, fromCurrency string) (float64, error)
}

type Service struct {
	groqAPIKey        string
	httpClient        *http.Client
	currencyConverter CurrencyConverter
}

func NewService(groqAPIKey string, currencyConverter CurrencyConverter) *Service {
	return &Service{
		groqAPIKey:        groqAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		currencyConverter: currencyConverter,
	}
}

type groqRequest struct {
	Messages []groqMessage `json:"messages"`
	Model    string        `json:"model"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *Service) GetAdvice(question string) (string, error) {
	if s.groqAPIKey == "" {
		return "", apperrors.ErrGroqAPIUnavailable
	}

	reqBody := groqRequest{
		Messages: []groqMessage{
			{
				Role:    "user",
				Content: question,
			},
		},
		Model: "llama-3.3-70b-versatile",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", apperrors.Wrap(err, "Ошибка сериализации запроса")
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", apperrors.Wrap(err, "Ошибка создания запроса")
	}

	req.Header.Set("Authorization", "Bearer "+s.groqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", apperrors.ErrGroqAPIUnavailable
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.Wrap(err, "Ошибка чтения ответа")
	}

	if resp.StatusCode != http.StatusOK {
		return "", apperrors.NewWithDetails(503, "Groq API недоступен", fmt.Sprintf("status: %d, body: %s", resp.StatusCode, string(body)))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", apperrors.Wrap(err, "Ошибка десериализации ответа")
	}

	if groqResp.Error != nil {
		return "", apperrors.NewWithDetails(503, "Ошибка от Groq", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return "", apperrors.New(503, "Модель не вернула текст ответа")
	}

	answer := groqResp.Choices[0].Message.Content
	if answer == "" {
		return "Модель не вернула текст ответа.", nil
	}

	return answer, nil
}

// AnalyzeFinances анализирует финансовую ситуацию пользователя
func (s *Service) AnalyzeFinances(req AnalysisRequest) (AnalysisResponse, error) {
	if s.groqAPIKey == "" {
		return AnalysisResponse{}, apperrors.ErrGroqAPIUnavailable
	}

	// Формируем промпт с инструкциями для ИИ
	additional := ""
	if req.Additional != nil && *req.Additional != "" {
		additional = "\n\nДополнительная информация: " + *req.Additional
	}

	prompt := fmt.Sprintf(`Ты финансовый консультант для российского рынка. Пользователь из России. Проанализируй финансовую ситуацию и дай конкретные рекомендации с учетом реалий РФ.

Данные пользователя (РФ):
- Статус: %s
- Ежемесячные расходы: %s
- Ежемесячные доходы: %s%s

Задача:
1. Извлеки из текста все суммы доходов и расходов (в рублях)
2. Посчитай общий месячный доход
3. Посчитай общие месячные расходы
4. Вычисли разницу (профицит или дефицит)
5. Дай конкретный финансовый совет с учетом российского рынка, законодательства РФ и экономической ситуации

Учитывай:
- Российские банки, вклады (ставки ЦБ РФ)
- Налоговое законодательство РФ (НДФЛ, налоговые вычеты)
- Российские финансовые инструменты (брокерские счета, ИИС, ОФЗ)
- Реалии российского рынка труда и социальной поддержки

СТРОГО верни ответ в таком формате (используй эти маркеры ТОЧНО):

===BALANCE===
Доход: X руб/мес
Расход: Y руб/мес
Профицит/Дефицит: Z руб/мес

===ADVICE===
[здесь конкретные рекомендации для российского рынка с учетом ситуации пользователя]

Не добавляй ничего лишнего. Используй маркеры ===BALANCE=== и ===ADVICE=== ТОЧНО как указано.`, req.Status, req.Expenses, req.Income, additional)

	// Отправляем запрос в Groq
	reqBody := groqRequest{
		Messages: []groqMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model: "llama-3.3-70b-versatile",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return AnalysisResponse{}, apperrors.Wrap(err, "Ошибка сериализации запроса")
	}

	httpReq, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return AnalysisResponse{}, apperrors.Wrap(err, "Ошибка создания запроса")
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.groqAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return AnalysisResponse{}, apperrors.ErrGroqAPIUnavailable
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AnalysisResponse{}, apperrors.Wrap(err, "Ошибка чтения ответа")
	}

	if resp.StatusCode != http.StatusOK {
		return AnalysisResponse{}, apperrors.NewWithDetails(503, "Groq API недоступен", fmt.Sprintf("status: %d", resp.StatusCode))
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return AnalysisResponse{}, apperrors.Wrap(err, "Ошибка десериализации ответа")
	}

	if groqResp.Error != nil {
		return AnalysisResponse{}, apperrors.NewWithDetails(503, "Ошибка от Groq", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return AnalysisResponse{}, apperrors.New(503, "Модель не вернула текст ответа")
	}

	answer := groqResp.Choices[0].Message.Content
	if answer == "" {
		return AnalysisResponse{}, apperrors.New(503, "Модель вернула пустой ответ")
	}

	// Парсим ответ (ищем БАЛАНС: и СОВЕТ:)
	return parseAnalysisResponse(answer), nil
}

// parseAnalysisResponse извлекает баланс и совет из ответа ИИ
func parseAnalysisResponse(text string) AnalysisResponse {
	// Ищем маркеры с помощью strings.Split
	balanceMarker := "===BALANCE==="
	adviceMarker := "===ADVICE==="

	balance := ""
	advice := ""

	// Разбиваем текст по маркерам
	if strings.Contains(text, balanceMarker) && strings.Contains(text, adviceMarker) {
		parts := strings.Split(text, balanceMarker)
		if len(parts) > 1 {
			afterBalance := parts[1]
			adviceParts := strings.Split(afterBalance, adviceMarker)
			
			if len(adviceParts) > 1 {
				balance = strings.TrimSpace(adviceParts[0])
				advice = strings.TrimSpace(adviceParts[1])
			}
		}
	}

	// Если парсинг не сработал, возвращаем весь текст как совет
	if balance == "" && advice == "" {
		return AnalysisResponse{
			Balance: "Данные недоступны",
			Advice:  strings.TrimSpace(text),
		}
	}

	return AnalysisResponse{
		Balance: balance,
		Advice:  advice,
	}
}

// GetStructuredAdvice обрабатывает структурированный запрос с конвертацией валют
func (s *Service) GetStructuredAdvice(ctx context.Context, req StructuredAdviceRequest) (*StructuredAdviceResponse, error) {
	if s.groqAPIKey == "" {
		return nil, apperrors.ErrGroqAPIUnavailable
	}

	// Конвертируем все доходы в рубли
	totalIncomeRUB := 0.0
	incomeDetails := []string{}
	for _, source := range req.IncomeSources {
		if source.Amount <= 0 {
			continue
		}
		
		amountInRUB, err := s.currencyConverter.ConvertToRUB(ctx, source.Amount, source.Currency)
		if err != nil {
			return nil, apperrors.Wrap(err, "Ошибка конвертации валюты")
		}
		
		totalIncomeRUB += amountInRUB
		incomeDetails = append(incomeDetails, fmt.Sprintf("%s: %.2f ₽ (из %.2f %s)", 
			getIncomeTypeLabel(source.Type), amountInRUB, source.Amount, source.Currency))
	}

	// Конвертируем все расходы в рубли
	totalExpensesRUB := 0.0
	expenseDetails := []string{}
	for _, source := range req.ExpenseSources {
		if source.Amount <= 0 {
			continue
		}
		
		amountInRUB, err := s.currencyConverter.ConvertToRUB(ctx, source.Amount, source.Currency)
		if err != nil {
			return nil, apperrors.Wrap(err, "Ошибка конвертации валюты")
		}
		
		totalExpensesRUB += amountInRUB
		expenseDetails = append(expenseDetails, fmt.Sprintf("%s: %.2f ₽ (из %.2f %s)", 
			getExpenseTypeLabel(source.Type), amountInRUB, source.Amount, source.Currency))
	}

	balance := totalIncomeRUB - totalExpensesRUB

	// Формируем промпт для AI
	question := buildFinancePrompt(
		totalIncomeRUB, 
		totalExpensesRUB, 
		balance, 
		incomeDetails, 
		expenseDetails, 
		req.Problems, 
		req.CustomProblem, 
		req.AdditionalInfo,
	)

	// Отправляем в Groq
	answer, err := s.GetAdvice(question)
	if err != nil {
		return nil, err
	}

	return &StructuredAdviceResponse{
		Answer:           answer,
		TotalIncomeRUB:   totalIncomeRUB,
		TotalExpensesRUB: totalExpensesRUB,
		BalanceRUB:       balance,
	}, nil
}

// buildFinancePrompt создает промпт для AI на основе структурированных данных
func buildFinancePrompt(
	totalIncome, totalExpenses, balance float64,
	incomeDetails, expenseDetails []string,
	problems []string,
	customProblem, additionalInfo string,
) string {
	var prompt strings.Builder

	prompt.WriteString("Ты — опытный финансовый советник, который понимает проблемы людей с небольшим доходом. ")
	prompt.WriteString("Говори просто, по-человечески, с заботой и без осуждения. Помоги этому человеку найти выход.\n\n")

	prompt.WriteString("**Откуда приходят деньги (всё конвертировано в рубли):**\n")
	for _, detail := range incomeDetails {
		prompt.WriteString(detail + "\n")
	}
	prompt.WriteString(fmt.Sprintf("**ИТОГО доход: %.2f ₽/мес**\n\n", totalIncome))

	prompt.WriteString("**Куда уходят деньги (всё конвертировано в рубли):**\n")
	for _, detail := range expenseDetails {
		prompt.WriteString(detail + "\n")
	}
	prompt.WriteString(fmt.Sprintf("**ИТОГО расход: %.2f ₽/мес**\n\n", totalExpenses))

	// Эмпатичное реагирование на баланс
	if balance < 0 {
		prompt.WriteString(fmt.Sprintf("**⚠️ ВАЖНО:** Человек сейчас в минусе (дефицит %.2f ₽). Ему ОЧЕНЬ тяжело.\n", -balance))
		prompt.WriteString("**Начни ответ с искреннего сочувствия и поддержки.** Признай что ситуация сложная, ")
		prompt.WriteString("скажи что понимаешь как это выматывает. Покажи что ты на его стороне. ")
		prompt.WriteString("Потом переходи к конкретным шагам выхода.\n\n")
	} else if balance > 0 && balance < totalIncome*0.15 {
		prompt.WriteString(fmt.Sprintf("**💪 Важный момент:** У человека небольшой плюс (остаётся %.2f ₽). Это РЕАЛЬНО здорово!\n", balance))
		prompt.WriteString("**Обязательно похвали** в начале ответа. Скажи что он молодец. Поддержи и мотивируй продолжать.\n\n")
	} else if balance >= totalIncome*0.15 {
		prompt.WriteString(fmt.Sprintf("**🎉 Отличная новость:** У человека хороший остаток (%.2f ₽)! Это достойный результат.\n", balance))
		prompt.WriteString("**Похвали и вдохнови** в начале. Он справляется лучше чем многие.\n\n")
	}

	if len(problems) > 0 {
		prompt.WriteString("**Что давит больше всего:**\n")
		for _, problem := range problems {
			prompt.WriteString(fmt.Sprintf("- %s\n", getProblemLabel(problem)))
		}
		prompt.WriteString("\n")
	}

	if customProblem != "" {
		prompt.WriteString(fmt.Sprintf("**В своих словах:** %s\n\n", customProblem))
	}

	if additionalInfo != "" {
		prompt.WriteString(fmt.Sprintf("**Дополнительно:** %s\n\n", additionalInfo))
	}

	prompt.WriteString("---\n\n")
	prompt.WriteString("Твоя задача:\n")
	prompt.WriteString("1. **Начни с поддержки.** Признай, что ситуация сложная, но выход есть.\n")
	prompt.WriteString("2. **Анализ без цифр и терминов.** Объясни простым языком, что происходит.\n")
	prompt.WriteString("3. **Конкретные шаги.** Дай 3-5 реальных действий, которые можно сделать прямо сейчас.\n")
	prompt.WriteString("4. **Говори \"вы\", \"вам\", \"можете\".** Как друг, который искренне хочет помочь.\n")
	prompt.WriteString("5. **Без финансового жаргона.** Вместо \"дефицит бюджета\" — \"денег не хватает\".\n")
	prompt.WriteString("6. **Надежда.** Покажи, что даже с таким доходом можно улучшить ситуацию.\n\n")
	prompt.WriteString("Формат ответа: обычный текст с разделением на абзацы. Используй жирный текст (**важное**) и списки где нужно.")

	return prompt.String()
}

// Вспомогательные функции для получения меток
func getIncomeTypeLabel(t string) string {
	labels := map[string]string{
		"salary":         "💼 Зарплата",
		"pension":        "👴 Пенсия",
		"bonus":          "🎁 Премии",
		"business":       "🏢 Бизнес/фриланс",
		"rental":         "🏠 Аренда",
		"children_help":  "👨‍👩‍👧 Помощь от близких",
		"investments":    "📈 Инвестиции",
		"other":          "📦 Другое",
	}
	if label, ok := labels[t]; ok {
		return label
	}
	return t
}

func getExpenseTypeLabel(t string) string {
	labels := map[string]string{
		"food":      "🍔 Еда",
		"utilities": "💡 Коммуналка",
		"credit":    "💳 Кредиты",
		"debt":      "📝 Долги",
		"transport": "🚗 Транспорт",
		"health":    "🏥 Здоровье",
		"general":   "📊 Бытовое",
		"other":     "📦 Другое",
	}
	if label, ok := labels[t]; ok {
		return label
	}
	return t
}

func getProblemLabel(p string) string {
	labels := map[string]string{
		"debt":       "💳 Долги душат",
		"budgeting":  "📅 До зарплаты не дотягиваю",
		"expenses":   "💸 Деньги утекают",
		"savings":    "💰 Хочу откладывать",
		"emergency":  "😰 Боюсь ЧП",
		"income":     "📉 Мало денег",
		"retirement": "👴 Страшно за будущее",
		"investing":  "📈 Хочу инвестировать",
	}
	if label, ok := labels[p]; ok {
		return label
	}
	return p
}

