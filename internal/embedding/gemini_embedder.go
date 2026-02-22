package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-ingestion/internal/domain/document"
	"strings" // <--- Added this import
	"time"
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

	modelName := "models/text-embedding-004"
	batchSize := 100
	sleepTime := 2 * time.Second 

	var allRecords []document.VectorRecord
	
	fmt.Printf("--- Embedding %d chunks (Batch Size: %d) ---\n", len(chunks), batchSize)

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		
		rawBatch := chunks[i:end]
		
		// 1. FILTER: Create a clean batch without empty strings
		var validBatch []document.Chunk
		var requests []batchEmbedRequest

		for _, chunk := range rawBatch {
			// If text is just whitespace, SKIP IT. Google will reject it.
			if strings.TrimSpace(chunk.Text) == "" {
				continue
			}
			
			validBatch = append(validBatch, chunk)
			requests = append(requests, batchEmbedRequest{
				Model: modelName,
				Content: content{
					Parts: []part{{Text: chunk.Text}},
				},
			})
		}

		// If the entire batch was empty lines, skip to next
		if len(requests) == 0 {
			continue
		}

		fmt.Printf("Processing batch %d to %d (Valid chunks: %d)...\n", i, end, len(validBatch))

		payload := batchEmbedPayload{Requests: requests}
		jsonBody, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal failed: %w", err)
		}

		// 2. Send Request
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/%s:batchEmbedContents", modelName)
		req, err := http.NewRequestWithContext(ctx, "POST", url+"?key="+e.apiKey, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("network error: %w", err)
		}
		
		// 3. Handle Errors
		if resp.StatusCode != http.StatusOK {
			var errBody map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errBody)
			resp.Body.Close()
			return nil, fmt.Errorf("api error (status %s): %v", resp.Status, errBody)
		}

		var response batchEmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode failed: %w", err)
		}
		resp.Body.Close()

		// 4. Map Results
		// IMPORTANT: We map 'response.Embeddings' to 'validBatch' (not rawBatch)
		// because we might have skipped some empty chunks.
		if len(response.Embeddings) != len(validBatch) {
			return nil, fmt.Errorf("mismatch: sent %d valid chunks but got %d vectors", len(validBatch), len(response.Embeddings))
		}

		for j, embedding := range response.Embeddings {
			allRecords = append(allRecords, document.VectorRecord{
				ID:       validBatch[j].ID,
				Vector:   embedding.Values,
				Chunk:    validBatch[j],
				Metadata: validBatch[j].Metadata,
			})
		}

		if end < len(chunks) {
			time.Sleep(sleepTime)
		}
	}

	return allRecords, nil
}

// --- STRUCTS (Keep these the same) ---
type batchEmbedPayload struct {
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