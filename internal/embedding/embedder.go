package embedding

import (
	"context"
	"rag-ingestion/internal/domain/document"
)

type Embedder interface {
	Embed(ctx context.Context, chunks []document.Chunk) ([]document.VectorRecord, error)
}