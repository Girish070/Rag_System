package implementation

import (
	"bufio"
	"bytes"
	"strings"

	"rag-ingestion/internal/domain/document"
)

type CodeParser struct {
	Language string
}

func NewCodeParser(language string) *CodeParser {
	return &CodeParser{
		Language: language,
	}
}

func (p *CodeParser) Parse(raw []byte) (*document.Document, error) {
	doc := &document.Document{
		ID:     generateDocumentID(),
		Blocks: []document.Block{},
		Metadata: map[string]string{
			"type":     "code",
			"language": p.Language,
		},
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))

	var buffer []string
	var inBlock bool

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if isFunctionOrClass(trimmed) {
			if inBlock {
				doc.Blocks = append(doc.Blocks, document.Block{
					Type:    document.BlockCode,
					Content: strings.Join(buffer, "\n"),
				})
				buffer = nil
			}
			inBlock = true
			buffer = append(buffer, line)
			continue
		}
		if inBlock {
			buffer = append(buffer, line)

			if isBlockEnd(trimmed) {
				doc.Blocks = append(doc.Blocks, document.Block{
					Type:    document.BlockCode,
					Content: strings.Join(buffer, "\n"),
				})
				buffer = nil
				inBlock = false
			}
			continue
		}
		doc.Blocks = append(doc.Blocks, document.Block{
			Type:    document.BlockCode,
			Content: line,
		})
		if len(buffer) > 0 {
			doc.Blocks = append(doc.Blocks, document.Block{
				Type:    document.BlockCode,
				Content: strings.Join(buffer, "\n"),
			})
		}
	}
	return doc, scanner.Err()
}

func isFunctionOrClass(line string) bool {
	keywords := []string{
		"func ",
		"def ",
		"class ",
		"public ",
		"private ",
		"protected ",
	}
	for _, kw := range keywords {
		if strings.HasPrefix(line, kw) {
			return true
		}
	}
	return false
}

func isBlockEnd(line string) bool {
	return line == "}" || strings.HasPrefix(line, "end")
}
