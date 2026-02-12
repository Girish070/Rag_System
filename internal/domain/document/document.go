package document

type Document struct {
	ID string
	Title string
	Blocks []Block
	Metadata map[string]string
}