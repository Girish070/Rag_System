package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-ingestion/internal/domain/document"
	"strings"
	"time"
)

type GeminiGenerator struct {
	url   string
	model string
}

// FIX 1: Fixed typo "Geberator" -> "Generator"
func NewGeminiGenerator() *GeminiGenerator {
	return &GeminiGenerator{
		url:   "http://localhost:11434/api/chat",
		model: "qwen2.5:0.5b",
	}
}

func (g *GeminiGenerator) GenerateAnswer(query string, ctxChunk []document.Chunk) (string, error) {
	// 1. Build the Context String
	var contextBuilder strings.Builder
	for _, Chunk := range ctxChunk {
		contextBuilder.WriteString(fmt.Sprintf("---\nFile: %s\nContent:\n%s\n", Chunk.Metadata["filename"], Chunk.Text))
	}

	// 2. Prepare the messages
	systemPrompt := "You are a helpful assistant. Answer the question strictly using the provided context. If the context contains code, explain it. If it contains text, summarize it."
	userMessage := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextBuilder.String(), query)

	payload := map[string]interface{}{
		"model":  g.model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
	}

	jsonBytes, _ := json.Marshal(payload)

	// 3. Send to Ollama (Localhost)
	client := &http.Client{Timeout: 120 * time.Second}

	// FIX 2: Changed "g.modelUrl" -> "g.url" to match the struct field
	req, _ := http.NewRequest("POST", g.url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama error: %s", resp.Status)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

func (g *GeminiGenerator) ImproveQuery(userQuery string) (string, error) {
	systemPrompt := "You are a helpful assistant. Answer the question strictly using the provided context. If the context contains code, explain it. If it contains text, summarize it."

	payload := map[string]interface{}{
		"model":  g.model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userQuery},
		},
	}
	jsonByte, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequest("POST", g.url, bytes.NewBuffer(jsonByte))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: AI Raw Response: '%s'\n", result.Message.Content)

	cleanQuery := strings.TrimSpace(strings.ReplaceAll(result.Message.Content, "/", ""))
	return cleanQuery, nil
}
