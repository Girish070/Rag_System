package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rag-ingestion/internal/domain/document"
)

const (
	// Fix 1: Corrected typo in variable name
	geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/embedding-001:batchEmbedContents"
)

type GeminiEmbedder struct {
	apiKey string
	client *http.Client
}

func NewGeminiEmbedder(apiKey string) *GeminiEmbedder {
	return &GeminiEmbedder{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (e *GeminiEmbedder) Embed(ctx context.Context, chunks []document.Chunk) ([]document.VectorRecord, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	// Prepare requests
	requests := make([]batchEmbedRequest, len(chunks))
	for i, chunk := range chunks {
		requests[i] = batchEmbedRequest{
			Model: "models/embedding-001",
			Content: content{
				Parts: []part{
					{Text: chunk.Text},
				},
			},
		}
	}

	// Fix 2: Field name in struct below must be 'Requests' to match JSON
	payload := batchEmbedPayload{Requests: requests}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Fix 3: Parameter is 'key', not 'keys'
	req, err := http.NewRequestWithContext(ctx, "POST", geminiURL+"?key="+e.apiKey, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// Fix 4: Content-Type must be 'application/json' (slash, not hyphen)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Tip: Decode the error body to see why Google rejected it
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("gemini api returned status: %s, details: %v", resp.Status, errResp)
	}

	var response batchEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Embeddings) != len(chunks) {
		return nil, errors.New("mismatch between number of chunks and number of vectors returned")
	}

	var records []document.VectorRecord
	for i, embedding := range response.Embeddings {
		records = append(records, document.VectorRecord{
			ID:       chunks[i].ID,
			Vector:   embedding.Values,
			Chunk:    chunks[i],
			Metadata: chunks[i].Metadata,
		})
	}

	return records, nil
}

// --- Internal Structs ---

type batchEmbedPayload struct {
	// Fix 5: JSON tag must be "requests" (plural)
	Requests []batchEmbedRequest `json:"requests"`
}

type batchEmbedRequest struct {
	Model   string  `json:"model"`
	Content content `json:"content"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type batchEmbedResponse struct {
	Embeddings []embeddingResult `json:"embeddings"`
}

type embeddingResult struct {
	Values []float32 `json:"values"`
}