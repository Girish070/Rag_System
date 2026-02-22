package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"rag-ingestion/internal/domain/document"
	"strings"
	"time"
)

type GeminiGenerator struct {
	apiKey string
	url    string
}

func NewGeminiGenerator() *GeminiGenerator {
	// 1. Ensure this matches your .env file exactly
	apiKey := os.Getenv("GEMINI_API_KEY") 
	return &GeminiGenerator{
		apiKey: apiKey,
		url:    "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent",
	}
}

func (g *GeminiGenerator) GenerateAnswer(query string, ctxChunk []document.Chunk) (string, error) {
	var contextBuilder strings.Builder
	for _, Chunk := range ctxChunk {
		// FIX #1: Use "filename" (what main.go sets), not "file"
		contextBuilder.WriteString(fmt.Sprintf("---\nFile: %s\nContent:\n%s\n", Chunk.Metadata["filename"], Chunk.Text))
	}

	// FIX #2: Added () to contextBuilder.String() so it sends TEXT, not a memory address
	userMessage := fmt.Sprintf(`CONTEXT FROM CODEBASE:
	%s
	USER QUESTION: %s`, contextBuilder.String(), query)

	systemInstruction := `You are an expert Go developer assisting with a codebase.
	INSTRUCTIONS:
	- Answer the question based ONLY on the provided context.
	- If the code snippet is provided, explain how it works.
	- Keep the answer concise, technical, and accurate.
	- Do not make up information not present in the context.`

	reqBody := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": systemInstruction},
			},
		},
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": userMessage},
				},
			},
		},
	}
	jsonBytes, _ := json.Marshal(reqBody)
	
	// Check if API Key is loaded
	if g.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is missing in .env")
	}

	req, _ := http.NewRequest("POST", g.url+"?key="+g.apiKey, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err // FIX #3: Return the actual error, don't return nil!
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned status: %s", resp.Status)
	}

	// FIX #4: Fixed struct tags ("candidates" vs "condidate")
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"` // <--- THIS WAS THE TYPO
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	// Check Candidates (Plural)
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "No answer generated", nil
}