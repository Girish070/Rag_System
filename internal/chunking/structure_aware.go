package chunking

import (
	"errors"
	//"fmt"
	"rag-ingestion/internal/domain/document"
	"strconv"
	"strings"
)

type StructureAwareChunker struct {
	MaxTokens     int
	OverlapTokens int
}

func NewStructureAwareChunker(maxTokens, overlapTokens int) *StructureAwareChunker {
	return &StructureAwareChunker{
		MaxTokens:     maxTokens,
		OverlapTokens: overlapTokens,
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

	chunks = applyOverlap(chunks, c.OverlapTokens)

	return chunks, nil
}

func newChunks(doc *document.Document, content string, index int) document.Chunk {
	metadata := make(map[string]string)

	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	metadata["chunk_index"] = strconv.Itoa(index)

	return document.Chunk{
		ID:         generateChunkID(doc.ID, index),
		Text:       content,
		DocumentID: doc.ID,
		Index:      index,
	}
}

func generateChunkID(docID string, index int) string {
	return docID + "_chunk_" + strconv.Itoa(index)
}

func estimateTokens(text string) int {
	return len(text) / 4
}

func applyOverlap(chunks []document.Chunk, overlapTokens int) []document.Chunk {

	if overlapTokens <= 0 || len(chunks) < 2 {
		return chunks
	}

	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1]
		curr := &chunks[i]

		if curr.Metadata["type"] == "code" {
			continue
		}

		overlapText := tailByToken(prev.Text, overlapTokens)
		if overlapText == "" {
			continue
		}

		curr.Text = overlapText + "\n" + curr.Text
		curr.Metadata["overlap"] = "true"
	}
	return chunks
}

func tailByToken(text string, maxTokens int) string {
	tokens := strings.Fields(text)

	if len(tokens) <= maxTokens {
		return text
	}

	start := len(tokens) - maxTokens
	return strings.Join(tokens[start:], "")
}
