package objects

import (
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/antoniszczepanik/gggit/utils"
)

type Object interface {
	GetContent() ([]byte, error)
	GetType() string
}

// Write object contents to internal git storage.
func Write(o Object) error {
	if err := IsEmpty(o); err != nil {
		return err
	}
	objectDir, err := utils.GetGitSubdir("objects")
	if err != nil {
		return err
	}
	hash, err := GetHash(o)
	if err != nil {
		return err
	}
	objectSubDir, objectName, err := utils.SplitHash(hash)
	if err != nil {
		return err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	// Create a subdirectory if does not exist.
	if _, err := os.ReadDir(objectSubDirPath); os.IsNotExist(err) {
		err = os.Mkdir(objectSubDirPath, 0755)
		if err != nil {
			return err
		}
	}
	objectFileName := filepath.Join(objectSubDirPath, objectName)

	// Skip if file already exists.
	if _, err := os.Stat(objectFileName); os.IsExist(err) {
		return nil
	}
	// Otherwise create a file.
	f, err := os.Create(objectFileName)
	if err != nil {
		return err
	}
	// Compress and write file contents.
	w := zlib.NewWriter(f)
	defer w.Close()
	fullContent, err := constructFullContent(o)
	if err != nil {
		return err
	}
	_, err = w.Write(fullContent)
	if err != nil {
		return err
	}
	return nil
}

// Read object object :).
func Read(hash string) (Object, error) {
	content, objectType, err := getObjectContent(hash)
	if err != nil {
		return nil, err
	}
	return parse(objectType, content)
}

// Print object contents by hash name.
func PrintObject(hash string) error {
	content, _, err := getObjectContent(hash)
	if err != nil {
		return err
	}
	fmt.Print(string(content))
	return nil
}

func getObjectContent(hash string) ([]byte, string, error) {
	objectDir, err := utils.GetGitSubdir("objects")
	if err != nil {
		return nil, "", err
	}
	objectSubDir, objectName, err := utils.SplitHash(hash)
	if err != nil {
		return nil, "", err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	if _, err := os.Stat(objectSubDirPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("object %s does not exist", hash)
	}

	objectPath := filepath.Join(objectSubDirPath, objectName)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("object %s does not exist", hash)
	}
	f, err := os.Open(objectPath)
	if err != nil {
		return nil, "", err
	}
	r, err := zlib.NewReader(f)
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	fullContents, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	pos := utils.GetNullBytePos(fullContents)
	if len(fullContents) < pos {
		return nil, "", errors.New("invalid object format")
	}
	objectType, _, err := splitHeader(string(fullContents[:pos]))
	if err != nil {
		return nil, "", err
	}
	return fullContents[pos+1:], objectType, nil
}

// Get hash of an object.
func GetHash(o Object) (string, error) {
	if err := IsEmpty(o); err != nil {
		return "", err
	}
	fullContent, err := constructFullContent(o)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1.Sum(fullContent)), nil
}

// Construct header of an object.
func getHeader(o Object) ([]byte, error) {
	if err := IsEmpty(o); err != nil {
		return nil, err
	}
	content, err := o.GetContent()
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf("%s %d\000", o.GetType(), len(content))), nil
}

// Get full object contents, header and actual content.
func constructFullContent(o Object) ([]byte, error) {
	if err := IsEmpty(o); err != nil {
		return nil, err
	}
	header, err := getHeader(o)
	if err != nil {
		return nil, err
	}
	content, err := o.GetContent()
	if err != nil {
		return nil, err
	}
	return append(header, content...), nil
}

// Parse file header to get object type and its size.
func splitHeader(header string) (string, int, error) {
	words := strings.Fields(header)
	if len(words) != 2 {
		return "", 0, errors.New("invalid header format")
	}
	size, err := strconv.Atoi(words[1])
	if err != nil {
		return "", 0, err
	}
	return words[0], size, nil
}

// Return specific object for arbitrary type and contents.
func parse(objectType string, contents []byte) (Object, error) {
	switch objectType {
	case blob:
		return parseBlob(contents)
	case commit:
		return parseCommit(contents)
	case tree:
		return parseTree(contents)
	}
	return nil, fmt.Errorf("cannot parse object %s", objectType)
}

type EmptyObjectError struct {
	Msg     string
	Content []byte
	Type    string
}

func (o EmptyObjectError) Error() string {
	return fmt.Sprintf("%s\n%s\n%s\n", o.Msg, o.Type, string(o.Content))
}

// Check if object was properly initialized.
func IsEmpty(o Object) error {
	content, err := o.GetContent()
	if err != nil {
		return err
	}
	if content == nil {
		return &EmptyObjectError{
			Msg:     "empty object content",
			Type:    o.GetType(),
			Content: content,
		}
	}
	if t := o.GetType(); t == "" {
		return &EmptyObjectError{
			Msg:     "empty object type",
			Type:    "",
			Content: content,
		}
	}
	return nil
}

func Exists(hash string) error {
	objectDir, err := utils.GetGitSubdir("objects")
	if err != nil {
		return err
	}
	objectSubDir, objectName, err := utils.SplitHash(hash)
	if err != nil {
		return err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	if _, err := os.Stat(objectSubDirPath); os.IsNotExist(err) {
		return err
	}
	objectFileName := filepath.Join(objectSubDirPath, objectName)
	if _, err := os.Stat(objectFileName); os.IsNotExist(err) {
		return err
	}
	return nil
}
