package chunking

import (
	"errors"
	"fmt"
	"rag-ingestion/internal/domain/document"
)

type StructureAwareChunker struct {
	MaxTokens int
}

func NewStructureAwereChunker(maxTokens int) *StructureAwareChunker {
	return &StructureAwareChunker{
		MaxTokens: maxTokens,
	}
}

func (c *StructureAwareChunker) Split(doc *document.Document) ([]document.Chunk, error) {
	if doc == nil {
		return nil, errors.New("Document is nil")
	}
	var chunks []document.Chunk
	var buffer string
	chunkIndex := 0

	for _, block := range doc.Blocks {
		if block.Type == document.BlockHeading && buffer != "" {
			chunks = append(chunks, newChunks(doc, buffer, chunkIndex))
			buffer = ""
			chunkIndex++
		}

		if block.Type == document.BlockCode {
			if buffer != "" {
				chunks = append(chunks, newChunks(doc, buffer, chunkIndex))
				buffer = ""
				chunkIndex++
			}

			chunks = append(chunks, newChunks(doc, block.Content, chunkIndex))
			chunkIndex++
			continue
		}

		buffer += block.Content + "\n"

		if estimateTokens(buffer) >= c.MaxTokens {
			chunks = append(chunks, newChunks(doc, buffer, chunkIndex))
			buffer = ""
			chunkIndex++
		}
	}

	if buffer != "" {
		chunks = append(chunks, newChunks(doc, buffer, chunkIndex))
	}
	return chunks, nil
}

func newChunks(doc *document.Document, content string, index int) document.Chunk {
	return document.Chunk{
		ID:         generateChunkID(doc.ID, index),
		Text:       content,
		DocumentID: doc.ID,
		Index:      index,
	}
}

func generateChunkID(docID string, index int) string {
	return fmt.Sprintf("%s-chunk-%d", docID, index)
}

func estimateTokens(text string) int {
	return len(text) / 4
}
