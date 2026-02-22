package storage

import (
	"context"

	"rag-ingestion/internal/domain/document"
)

type VectorStore interface {
	Upsert(ctx context.Context, records []document.VectorRecord) error
	Search(ctx context.Context, queryVector []float32, limit int) ([]document.Chunk, error)
}
