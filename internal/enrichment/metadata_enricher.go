package enrichment

import "rag-ingestion/internal/domain/document"

type MetadataEnricher interface {
	Apply(chunks []document.Chunk) ([]document.Chunk, error)
}