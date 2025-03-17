package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Структуры для ответа от API
type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type ApiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Структура для входящего запроса с клиентской стороны
type PromptRequest struct {
	Prompt string `json:"prompt"`
}

// Функция, которая делает запрос к внешнему API с переданным prompt
func generateContent(prompt string) (string, error) {
	API_KEY := "AIzaSyBpCnsp5BCDCHylAdPaGy2yH8wGzuVC98M"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", API_KEY)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp ApiResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %v, response: %s", err, string(bodyBytes))
	}

	if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
		return apiResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("unexpected JSON structure")
}

// HTTP-обработчик для эндпоинта /generate
func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var reqData PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	if reqData.Prompt == "" {
		http.Error(w, "Поле prompt не может быть пустым", http.StatusBadRequest)
		return
	}

	result, err := generateContent(reqData.Prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при генерации контента: %v", err), http.StatusInternalServerError)
		return
	}

	respData := map[string]string{
		"response": result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respData)
}

func main() {
	http.HandleFunc("/generate", generateHandler)
	fmt.Println("Server start on port 9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
