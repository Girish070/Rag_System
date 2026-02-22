package retrieval

import (
	"context"
	"fmt"
	"rag-ingestion/internal/domain/document"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/storage"
)

type Retriever struct {
	embedder embedding.Embedder
	store    storage.VectorStore
}

func NewRetriever(embedder embedding.Embedder, store storage.VectorStore) *Retriever {
	return &Retriever{
		embedder: embedder,
		store:    store,
	}
}

func (r *Retriever) Query(ctx context.Context, queryText string, limit int) ([]document.Chunk, error) {
	queryChunk := document.Chunk{Text: queryText}

	vectors, err := r.embedder.Embed(ctx, []document.Chunk{queryChunk})
	if err != nil {
		return nil, fmt.Errorf("failed to Embed Query: %w", err)
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("Embedding returned no vectors")
	}

	results, err := r.store.Search(ctx, vectors[0].Vector, limit)
	if err != nil {
		return nil, fmt.Errorf("storage search failed: %w", err)
	}
	return results, nil
}
