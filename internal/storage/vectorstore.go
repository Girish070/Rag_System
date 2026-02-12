package storage

import (
	"context"

	"rag-ingestion/internal/domain/document"
)

type VectorStore interface {
	Upsert(ctx context.Context, records []document.VectorRecord) error
}
