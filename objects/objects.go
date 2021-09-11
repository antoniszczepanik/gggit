package objects

import (
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antoniszczepanik/gggit/utils"
)

const HeaderFmt = "%s %d\000"

type Object interface {
	GetContent() (string, error)
	GetType() string
	Write() error
}

// Generic write object method.
func Write(o Object) error {
	if err := IsEmpty(o); err != nil {
		return err
	}
	objectDir, err := utils.GetGitSubdir("objects")
	if err != nil {
		return err
	}
	hash, err := CalculateHash(o)
	if err != nil {
		return err
	}
	objectSubDir, objectName, err := utils.SplitHash(hash)
	if err != nil {
		return err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
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
	rawContent, err := constructRawContent(o)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(rawContent))
	if err != nil {
		return err
	}
	return nil
}

func Read(hash string) (Object, error) {
	rawContent, err := getObjectRawContent(hash)
	if err != nil {
		return nil, err
	}
	objType, _, content, err := splitRawContent(rawContent)
	if err != nil {
		return nil, err
	}
	switch objType {
	case blob:
		return parseBlob(content)
	case tree:
		return parseTree(content)
	case commit:
		return parseCommit(content)
	default:
		return nil, fmt.Errorf("unexpected object type %s", objType)
	}
}

// Print object contents by hash name.
func PrintObject(hash string) error {
	rawContent, err := getObjectRawContent(hash)
	if err != nil {
		return err
	}
	_, _, content, err := splitRawContent(rawContent)
	if err != nil {
		return err
	}
	fmt.Print(content)
	return nil
}

func getObjectRawContent(hash string) (string, error) {
	objectDir, err := utils.GetGitSubdir("objects")
	if err != nil {
		return "", err
	}
	objectSubDir, objectName, err := utils.SplitHash(hash)
	if err != nil {
		return "", err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	if _, err := os.Stat(objectSubDirPath); os.IsNotExist(err) {
		return "", fmt.Errorf("object %s does not exist", hash)
	}

	objectPath := filepath.Join(objectSubDirPath, objectName)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return "", fmt.Errorf("object %s does not exist", hash)
	}
	f, err := os.Open(objectPath)
	if err != nil {
		return "", err
	}
	r, err := zlib.NewReader(f)
	if err != nil {
		return "", err
	}
	defer r.Close()
	rawContent, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(rawContent), nil
}

// Calculate hash of an object.
func CalculateHash(o Object) (string, error) {
	if err := IsEmpty(o); err != nil {
		return "", err
	}
	rawContent, err := constructRawContent(o)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(rawContent))), nil
}

// Construct header of an object.
func getHeader(o Object) (string, error) {
	if err := IsEmpty(o); err != nil {
		return "", err
	}
	content, err := o.GetContent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(HeaderFmt, o.GetType(), len(content)), nil
}

// Get raw object contents, header and actual content.
func constructRawContent(o Object) (string, error) {
	if err := IsEmpty(o); err != nil {
		return "", err
	}
	header, err := getHeader(o)
	if err != nil {
		return "", err
	}
	content, err := o.GetContent()
	if err != nil {
		return "", err
	}
	return header + content, nil
}

type EmptyObjectError struct {
	Msg     string
	Content string
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
	if content == "" {
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

// Split raw object content into object type, size and actual content.
func splitRawContent(rawContent string) (string, int, string, error) {
	headerEnd := strings.Index(rawContent, "\000")
	if headerEnd == -1 {
		return "", 0, "", errors.New("no null byte in raw content")
	}
	header, content := rawContent[:headerEnd+1], rawContent[headerEnd+1:]
	objType, objSize, err := parseHeader(header)
	if err != nil {
		return "", 0, "", err
	}
	return objType, objSize, content, nil
}

func parseHeader(header string) (string, int, error) {
	var (
		objType string
		objSize int
	)
	r := strings.NewReader(header)
	_, err := fmt.Fscanf(r, HeaderFmt, &objType, &objSize)
	if err != nil {
		return "", 0, err
	}
	return objType, objSize, nil
}
