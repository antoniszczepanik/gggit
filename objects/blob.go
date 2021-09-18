package objects

import (
	"os"
)

const BlobObject ObjectType = "blob"

type Blob struct {
	content string
}

func NewBlob(content string) Blob {
	return Blob{content: content}
}

func NewBlobFromFile(filepath string) (Blob, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return Blob{}, err
	}
	return NewBlob(string(content)), nil
}

func (b Blob) GetContent() (string, error) {
	return b.content, nil
}

func (b Blob) GetType() ObjectType {
	return BlobObject
}

func (b Blob) Write() error {
	return Write(b)
}

func parseBlob(content string) (Blob, error) {
	return NewBlob(content), nil
}
