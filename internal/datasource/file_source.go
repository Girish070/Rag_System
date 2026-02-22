package datasource

import (
	"io"
	"os"
	"path/filepath"
)

type FileSource struct {
	path string
}

func NewFileSource(path string) *FileSource {
	return &FileSource{
		path: path,
	}
}

func (fs *FileSource) Read() ([]byte, error) {
	file, err := os.Open(fs.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func (fs *FileSource) Metadata() map[string]string {
	return map[string]string{
		"source":   "local_file",
		"filename": filepath.Base(fs.path),
		"path":     fs.path,
	}
}
