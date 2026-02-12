package chunking

import (
	"errors"
	"rag-ingestion/internal/domain/document"
)

type StructureAwareChunker struct {
	MaxTokens int
}

func NewStructureAwareChunker(maxTokens int) *StructureAwareChunker {
	return &StructureAwareChunker{
		MaxTokens: maxTokens,
	}
}

func (c *StructureAwareChunker) Split(doc document.Document) ([]document.Chunk, error) {
	Chunker.Split(doc)
	if doc == nil {
		return nil, errors.New("Document is nil")
	}
	var chunks []document.Chunks
	var buffer string
	chunkIndex := 0

	for _, block := range doc.Block {
		if block.Type == document.BlockHeading && buffer != "" {
			chunks = append(chunks, newChunk(doc, buffer, chunkIndex))
			buffer = ""
			chunkIndex++
		}
		if block.Type == document.BlockCode {
			if buffer != "" {
				chunks = append(chunks, newChunk(doc, buffer, chunkIndex))
				buffer = ""
				chunkIndex++
			}
			chunks = append(chunks, newChunk(doc, buffer, chunkIndex))
			chunkIndex++
			continue
		}
		buffer += block.Content + "\n"
		if estimateTokens(buffer) >= c.MaxTokens {
			chunks = append(chunks, newChunk(doc, buffer, chunkIndex))
			buffer = ""
			chunkIndex++
		}
		if buffer != "" {
			chunks = append(chunks, newChunk(doc, buffer, chunkIndex))
		}
	}
	return chunks, nil
}

func newChunk(doc *document.Document, content string, index int) document.Chunk {
	return document.Chunk{
		ID:         generateChunkID(doc.ID, index),
		Text:       content,
		DocumentID: doc.ID,
		Index:      index,
	}
}

func estimateTokens(text string) int{
	return len(text) / 4
}