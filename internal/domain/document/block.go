package document

type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockHeading   BlockType = "heading"
	BlockCode      BlockType = "code"
	BlockList      BlockType = "type"
	BlockTable     BlockType = "table"
)

type Block struct {
	Type     BlockType
	Content  string
	Metadata map[string]string
}
