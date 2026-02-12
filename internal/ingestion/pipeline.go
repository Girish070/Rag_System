package ingestion

import (
	"context"
	"rag-ingestion/internal/chunking"
	"rag-ingestion/internal/embedding"
	"rag-ingestion/internal/enrichment"
	"rag-ingestion/internal/parser"
	"rag-ingestion/internal/storage"
)

type Pipeline struct {
	parser   parser.Parser
	chunker  chunking.Chunker
	enricher enrichment.MetadataEnricher
	embedder embedding.Embedder
	vectorDB storage.VectorStore
}

func NewPipeline(parser parser.Parser,
	chunker chunking.Chunker,
	enricher enrichment.MetadataEnricher,
	embedder embedding.Embedder,
	vectorDB storage.VectorStore,
) *Pipeline {
	return &Pipeline{
		parser:   parser,
		chunker:  chunker,
		enricher: enricher,
		embedder: embedder,
		vectorDB: vectorDB,
	}
}

func (p *Pipeline) Run(ctx context.Context, source datasource.DataSource) error {
	raw, err := source.Read()
	if err != nil {
		return err
	}
	doc, err := p.parser.Parse(raw)
	if err != nil {
		return err
	}

	doc.Metadata = mergeMetadata(
		doc.Metadata,
		source.Metadata(),
	)

	chunks, err := p.chunker.Split(doc)
	if err != nil {
		return err
	}

	enrichedChunks, err := p.enricher.Apply(chunks)
	if err != nil {
		return err
	}

	vectors, err := p.embedder.Embed(ctx, enrichedChunks)
	if err != nil {
		return err
	}

	if err := p.vectorDB.Upsert(ctx, vectors); err != nil {
		return err
	}
	return nil
}

func mergeMetadata(a, b map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
