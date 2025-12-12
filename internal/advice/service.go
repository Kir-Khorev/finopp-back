package advice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/Kir-Khorev/finopp-back/pkg/errors"
)

type Service struct {
	groqAPIKey string
	httpClient *http.Client
}

func NewService(groqAPIKey string) *Service {
	return &Service{
		groqAPIKey: groqAPIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

