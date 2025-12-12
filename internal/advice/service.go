package advice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
		return "", fmt.Errorf("API ключ Groq не настроен на сервере")
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
		return "", fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.groqAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к Groq API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Groq API вернул ошибку: %s (код %d)", string(body), resp.StatusCode)
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	if groqResp.Error != nil {
		return "", fmt.Errorf("ошибка от Groq: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("модель не вернула текст ответа")
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
		return AnalysisResponse{}, fmt.Errorf("API ключ Groq не настроен на сервере")
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
		return AnalysisResponse{}, fmt.Errorf("ошибка сериализации запроса: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.groqAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("ошибка запроса к Groq API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AnalysisResponse{}, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return AnalysisResponse{}, fmt.Errorf("Groq API вернул ошибку: %s (код %d)", string(body), resp.StatusCode)
	}

	var groqResp groqResponse
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return AnalysisResponse{}, fmt.Errorf("ошибка десериализации ответа: %w", err)
	}

	if groqResp.Error != nil {
		return AnalysisResponse{}, fmt.Errorf("ошибка от Groq: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return AnalysisResponse{}, fmt.Errorf("модель не вернула текст ответа")
	}

	answer := groqResp.Choices[0].Message.Content
	if answer == "" {
		return AnalysisResponse{}, fmt.Errorf("модель вернула пустой ответ")
	}

	// Парсим ответ (ищем БАЛАНС: и СОВЕТ:)
	return parseAnalysisResponse(answer), nil
}

// parseAnalysisResponse извлекает баланс и совет из ответа ИИ
func parseAnalysisResponse(text string) AnalysisResponse {
	balance := ""
	advice := ""

	// Простой парсинг по маркерам
	lines := splitLines(text)
	section := ""

	for _, line := range lines {
		trimmed := trim(line)
		
		if contains(trimmed, "===BALANCE===") {
			section = "balance"
			continue
		}
		
		if contains(trimmed, "===ADVICE===") {
			section = "advice"
			continue
		}

		if trimmed == "" {
			continue
		}

		if section == "balance" {
			if balance != "" {
				balance += "\n"
			}
			balance += trimmed
		} else if section == "advice" {
			if advice != "" {
				advice += "\n"
			}
			advice += trimmed
		}
	}

	// Если парсинг не сработал, возвращаем весь текст
	if balance == "" && advice == "" {
		return AnalysisResponse{
			Balance: "Не удалось извлечь данные",
			Advice:  text,
		}
	}

	return AnalysisResponse{
		Balance: trim(balance),
		Advice:  trim(advice),
	}
}

// Вспомогательные функции для парсинга
func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func trim(s string) string {
	start := 0
	end := len(s)
	
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

