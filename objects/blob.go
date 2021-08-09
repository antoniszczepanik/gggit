package objects

import (
	"os"
)

const blob string = "blob"

type Blob struct {
	fileData []byte
}

func (b Blob) GetContent() ([]byte, error) {
	return b.fileData, nil
}

func (b Blob) GetType() string {
	return blob
}

func parseBlob(contents []byte) (Blob, error) {
	return Blob{
		fileData: contents,
	}, nil
}

// Create blob object from contents of a file.
func CreateBlobObject(filepath string) (Blob, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return Blob{}, err
	}
	return Blob{
		fileData: content,
	}, nil
}
