package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rag-ingestion/internal/domain/document"
)

type OllamaEmbedder struct {
	url   string
	model string
}

func NewOllamaEmbedder() *OllamaEmbedder {
	return &OllamaEmbedder{
		url:   "http://localhost:11434/api/embed",
		model: "nomic-embed-text",
	}
}

type ollamaRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (e *OllamaEmbedder) Embed(ctx context.Context, chunks []document.Chunk) ([]document.VectorRecord, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	var inputs []string
	var validChunks []document.Chunk

	for _, chunk := range chunks {
		if chunk.Text != "" {
			inputs = append(inputs, chunk.Text)
			validChunks = append(validChunks, chunk)
		}
	}
	reqBody := ollamaRequest{
		Model: e.model,
		Input: inputs,
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := http.Post(e.url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("the ollama is ofline make sure its running: %v", err)
	}
	defer resp.Body.Close()

	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %v", err)
	}

	var records []document.VectorRecord
	for i, emb := range result.Embeddings {
		records = append(records, document.VectorRecord{
			ID:       validChunks[i].ID,
			Vector:   emb,
			Chunk:    validChunks[i],
			Metadata: validChunks[i].Metadata,
		})
	}
	return records, nil
}
