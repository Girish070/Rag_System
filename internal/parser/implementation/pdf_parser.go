package implementation

import (
	"bytes"
	"rag-ingestion/internal/domain/document"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type PdfParser struct{}

func NewPdfParser() *PdfParser {
	return &PdfParser{}
}

func (p *PdfParser) Parse(raw []byte) (*document.Document, error) {
	reader := bytes.NewReader(raw)
	pdfReader, err := pdf.NewReader(reader, int64(len(raw)))
	if err != nil {
		return nil, err
	}
	doc := document.Document{
		ID:     generateDocumentID(),
		Blocks: []document.Block{},
		Metadata: map[string]string{
			"type": "pdf",
		},
	}
	numPages := pdfReader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			text := strings.TrimSpace(line)

			if text == "" {
				continue
			}
			if looksLikeHeading(text) {
				doc.Blocks = append(doc.Blocks, document.Block{
					Type:    document.BlockHeading,
					Content: text,
					Metadata: map[string]string{
						"page": strconv.Itoa(i),
					},
				})
				continue
			}
			doc.Blocks = append(doc.Blocks, document.Block{
				Type:    document.BlockParagraph,
				Content: text,
				Metadata: map[string]string{
					"page": strconv.Itoa(i),
				},
			})
		}
	}
	return &doc, nil
}

func looksLikeHeading(text string) bool {
	if len(text) > 80 {
		return false
	}

	if text == strings.ToUpper(text) {
		return true
	}

	return !strings.HasSuffix(text, ".")
}

func generateDocumentID() string {
	return uuid.New().String()
}