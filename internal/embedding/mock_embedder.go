package embedding

import (
	"context"
	"rag-ingestion/internal/domain/document"
)

type MockEmbedder struct{}

func NewMockWmbedder() *MockEmbedder {
	return &MockEmbedder{}
}

func (e *MockEmbedder) Embed(ctx context.Context, chunks []document.Chunk) ([]document.VectorRecord, error) {
	var records []document.VectorRecord

	for _, chunk := range chunks {
		vector := make([]float32, 768)
		val := float32(len(chunk.Text)%100) / 100.0
		for j := 0; j < 768; j++ {
			if j%2 == 0 {
				vector[j] = val
			} else {
				vector[j] = 1.0 - val
			}
		}
		records = append(records, document.VectorRecord{
			ID:       chunk.ID,
			Vector:   vector,
			Chunk:    chunk,
			Metadata: chunk.Metadata,
		})
	}
	return records, nil
}
