package main

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
)

type Object interface {
	GetContent() ([]byte, error)
	GetType() string
}

type Blob struct {
	fileData []byte
}

func (b Blob) GetContent() ([]byte, error) {
	return b.fileData, nil
}

func (b Blob) GetType() string {
	return "blob"
}

type Tree struct {
}

func (t Tree) GetContent() ([]byte, error) {
	//TODO: Implement me.
	return make([]byte, 1), nil
}

func (t Tree) GetType() string {
	return "tree"
}

type Commit struct {
}

func (t Commit) GetContent() ([]byte, error) {
	//TODO: Implement me.
	return make([]byte, 1), nil
}

func (t Commit) GetType() string {
	return "commit"
}

// Write object contents to internal git storage.
func Write(o Object) error {
	if err := checkIfEmpty(o); err != nil {
		return err
	}
	objectDir, err := GetGitSubdir("objects")
	if err != nil {
		return err
	}
	hash, err := GetHash(o)
	if err != nil {
		return err
	}
	objectSubDir, objectName, err := splitHash(hash)
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

// Read internal git object by hash value.
func ReadObject(hash string) (Object, error) {
	objectDir, err := GetGitSubdir("objects")
	if err != nil {
		return nil, err
	}
	objectSubDir, objectName, err := splitHash(hash)
	if err != nil {
		return nil, err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	if _, err := os.Stat(objectSubDirPath); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("object %s does not exist", hash))
	}

	objectPath := filepath.Join(objectSubDirPath, objectName)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("object %s does not exist", hash))
	}
	f, err := os.Open(objectPath)
	if err != nil {
		return nil, err
	}
	r, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	fullContents, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	pos := getNullBytePos(fullContents)
	if len(fullContents) < pos {
		return nil, errors.New("invalid object format")
	}
	objectType, _, err := splitHeader(string(fullContents[:pos]))
	if err != nil {
		return nil, err
	}
	return parseObject(objectType, fullContents[pos+1:])
}

// Get directory and filesystem object name.
func splitHash(hash string) (string, string, error) {
	if len(hash) != 40 {
		return "", "", errors.New("incorrect hash length")
	}
	return hash[:2], hash[2:], nil
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

func parseObject(objectType string, contents []byte) (Object, error) {
	switch objectType {
	case "blob":
		return parseBlob(contents)
	case "commit":
		// TODO parseCommit(contents)
		return nil, errors.New("cannot parse commit objects")
	case "tree":
		// TODO parseTree(contents)
		return nil, errors.New("cannot parse tree objects")
	}
	return nil, errors.New(fmt.Sprintf("cannot parse object %s", objectType))
}

func parseBlob(contents []byte) (Blob, error) {
	return Blob{
		fileData: contents,
	}, nil
}

// Get hash of an object.
func GetHash(o Object) (string, error) {
	if err := checkIfEmpty(o); err != nil {
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
	if err := checkIfEmpty(o); err != nil {
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
	if err := checkIfEmpty(o); err != nil {
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

// Validate if object has any contents.
func checkIfEmpty(o Object) error {
	if content, err := o.GetContent(); err != nil || content == nil || o.GetType() == "" {
		return errors.New("object is not complete")
	}
	return nil
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

func CreateTreeObject(dirpath string) (*Tree, error) {
	// TODO
	fmt.Println(dirpath)
	return nil, errors.New("Not implemented")
}

func CreateCommitOboject(treehash string) (*Commit, error) {
	// TODO
	fmt.Println(treehash)
	return nil, errors.New("Not implemented")
}

// Get position of the null byte, to seperate header from gg
func getNullBytePos(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
