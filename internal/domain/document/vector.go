package document

type VectorRecord struct {
	ID       string
	Vector   []float32
	Chunk    Chunk
	Metadata map[string]string
}
