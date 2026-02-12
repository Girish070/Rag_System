package document

type Chunk struct {
	ID         string
	DocumentID string
	Index      int
	Text       string
	Metadata   map[string]string
}
