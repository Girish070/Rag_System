package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"rag-ingestion/internal/domain/document"
	"strings"
	"time"
)

type NvidiaGenerator struct {
	apiKey string
	client *http.Client
	model  string
}

func NewNvidiaGenerator() *NvidiaGenerator {
	return &NvidiaGenerator{
		apiKey: os.Getenv("NVIDIA_API_KEY"),
		client: &http.Client{Timeout: 30 * time.Second},
		model:  "meta/llama-3.1-8b-instruct", // Very fast, great for coding and RAG
	}
}

func (g *NvidiaGenerator) GenerateAnswer(query string, ctxChunk []document.Chunk) (string, error) {
	var contextBuilder strings.Builder
	for _, chunk := range ctxChunk {
		contextBuilder.WriteString(fmt.Sprintf("---\nFile: %s\nContent:\n%s\n", chunk.Metadata["filename"], chunk.Text))
	}

	systemPrompt := `You are an intelligent research assistant. Your job is to answer the user's question based ONLY on the provided context. 
	
	Rule 1: If the context provides the necessary information, formulate a clear, helpful, and beautifully formatted answer.
	Rule 2: If the provided context does NOT contain the answer to the user's question, you must reply with exactly this phrase and nothing else: 'I cannot answer'.`
	userPrompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", contextBuilder.String(), query)

	return g.callNvidiaAPI(systemPrompt, userPrompt)
}

func (g *NvidiaGenerator) ImproveQuery(userQuery string) (string, error) {
	systemPrompt := "You are a query extractor. Extract 3 to 5 technical keywords from the user's question. Return ONLY the keywords separated by spaces. Do not add any conversational text."
	
	resp, err := g.callNvidiaAPI(systemPrompt, userQuery)
	if err != nil {
		return "", err
	}
	
	cleanQuery := strings.TrimSpace(strings.ReplaceAll(resp, "\"", ""))
	return cleanQuery, nil
}

func (g *NvidiaGenerator) callNvidiaAPI(system string, user string) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY is not set in .env")
	}

	// Nvidia uses the standard OpenAI chat completions endpoint format
	url := "https://integrate.api.nvidia.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": g.model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"max_tokens": 1024,
	}

	jsonBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey) // Inject the Nvidia Key

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(bodyBytes))
	}

	// Parse the OpenAI-compatible response format
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("nvidia returned an empty response")
}