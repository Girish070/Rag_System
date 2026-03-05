package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type TavilyAgent struct {
	apiKey string
	client *http.Client
}

func NewTavilyAgent() *TavilyAgent {
	return &TavilyAgent{
		apiKey: os.Getenv("TAVILY_API_KEY"),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *TavilyAgent) SearchWeb(query string) (string, error) {
	if a.apiKey == "" {
		return "", fmt.Errorf("TAVILY_API_KEY is not set in .env")
	}

	// 1. Removed "api_key" from the JSON body
	reqBody := map[string]interface{}{
		"query":          query,
		"search_depth":   "basic",
		"include_answer": false,
		"max_results":    3,
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonBytes))
	
	// 2. Add the strict Security Headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey) // <--- THIS WAS MISSING!

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error hitting tavily: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tavily API error: %s", string(bodyBytes))
	}

	// Parse the response
	var tResp struct {
		Results []struct {
			Url     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(bodyBytes, &tResp); err != nil {
		return "", err
	}

	var contextBuilder strings.Builder
	for _, r := range tResp.Results {
		contextBuilder.WriteString(fmt.Sprintf("---\nSource URL: %s\nScraped Content:\n%s\n", r.Url, r.Content))
	}

	return contextBuilder.String(), nil
}