package parser

import "rag-ingestion/internal/domain/document"

type Parser interface {
	Parse(raw []byte) (*document.Document, error)
}