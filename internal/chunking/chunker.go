package chunking

import "rag-ingestion/internal/domain/document"

type Chunker interface {
	Split(doc *document.Document) ([]document.Chunk, error)
}