package enrichment

import "rag-ingestion/internal/domain/document"

type NoOpEnricher struct{}

func NewNoOpEnricher() *NoOpEnricher {
	return &NoOpEnricher{}
}

func (e *NoOpEnricher) Apply(chunks []document.Chunk) ([]document.Chunk, error) {
	return chunks, nil
}
