package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

type Tree []treeEntry

type treeEntry struct {
	Mode  string
	Hash  string
	Name  string
	Entry Object
}

const treeEntryFmt = "%s %s %s\t%s"

func (t Tree) GetContent() ([]byte, error) {
	var content []byte
	for _, e := range t {
		if e.Mode == "" || e.Hash == "" || e.Name == "" || e.Entry == nil {
			return nil, errors.New("cannot get content of tree with missing attributes")
		}
		entry_content := fmt.Sprintf(treeEntryFmt+"\n", e.Mode, e.Entry.GetType(), e.Hash, e.Name)
		content = append(content, []byte(entry_content)...)
	}
	return content, nil
}

func (t Tree) GetType() string {
	return "tree"
}

type Commit struct {
	TreeHash   string
	ParentHash string
	Author     AuthorType
	// TODO: Should be represented as in RFC 2822. On drive it will be stored
	// as unixts timezoneoffset
	Time time.Time
	Msg  string
}

type AuthorType struct {
	Name  string
	Email string
}

func (c Commit) GetContent() ([]byte, error) {
	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	if c.ParentHash != "" {
		content += fmt.Sprintf("parent %s\n", c.ParentHash)
	}
	_, offset := c.Time.Zone()
	time := fmt.Sprintf("%d %d", c.Time.Unix(), offset)
	content += fmt.Sprintf("author %s <%s> %s\n\n", c.Author.Name, c.Author.Email, time)
	content += fmt.Sprintf("%s\n", c.Msg)
	return []byte(content), nil
}

func (c Commit) GetType() string {
	return "commit"
}

// Write object contents to internal git storage.
func Write(o Object) error {
	if err := isEmpty(o); err != nil {
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

// Read object object :)
func ReadObject(hash string) (Object, error) {
	content, objectType, err := getObjectContent(hash)
	if err != nil {
		return nil, err
	}
	return parseObject(objectType, content)
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
	objectDir, err := GetGitSubdir("objects")
	if err != nil {
		return nil, "", err
	}
	objectSubDir, objectName, err := splitHash(hash)
	if err != nil {
		return nil, "", err
	}
	objectSubDirPath := filepath.Join(objectDir, objectSubDir)
	if _, err := os.Stat(objectSubDirPath); os.IsNotExist(err) {
		return nil, "", errors.New(fmt.Sprintf("object %s does not exist", hash))
	}

	objectPath := filepath.Join(objectSubDirPath, objectName)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		return nil, "", errors.New(fmt.Sprintf("object %s does not exist", hash))
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
	pos := getNullBytePos(fullContents)
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
	if err := isEmpty(o); err != nil {
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
	if err := isEmpty(o); err != nil {
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
	if err := isEmpty(o); err != nil {
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

// Return specific object for arbitrary type and contents.
func parseObject(objectType string, contents []byte) (Object, error) {
	switch objectType {
	case "blob":
		return parseBlob(contents)
	case "commit":
		// TODO parseCommit(contents)
		return nil, errors.New("cannot parse commit objects")
	case "tree":
		return parseTree(contents)
	}
	return nil, errors.New(fmt.Sprintf("cannot parse object %s", objectType))
}

func parseBlob(contents []byte) (Blob, error) {
	return Blob{
		fileData: contents,
	}, nil
}

func parseTree(contents []byte) (Tree, error) {
	var (
		t     Tree
		emode string
		etype string
		ehash string
		ename string
	)
	rawEntries := splitEntries(contents)
	for _, rawEntry := range rawEntries {
		r := bytes.NewReader(rawEntry)
		_, err := fmt.Fscanf(r, treeEntryFmt, &emode, &etype, &ehash, &ename)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Tree{}, err
		}
		obj, err := ReadObject(ehash)
		if err != nil {
			return Tree{}, err
		}
		t = append(t, treeEntry{Mode: emode, Hash: ehash, Name: ename, Entry: obj})
	}
	return t, nil
}

func splitEntries(contents []byte) [][]byte {
	var (
		splitted [][]byte
		split    []byte
	)
	for _, c := range contents {
		if c != '\n' {
			split = append(split, c)
		} else {
			splitted = append(splitted, split)
			split = nil
		}
	}
	return splitted
}

func parseCommit(contents []byte) (Commit, error) {
	return Commit{}, nil
}

// Write tree object, recursively writing all its entries.
func WriteTree(t Tree) error {
	if err := Write(t); err != nil {
		return err
	}
	for _, tEntry := range t {
		if exists(tEntry.Hash) == nil {
			continue
		}
		if tEntry.Entry.GetType() == "blob" {
			if err := Write(tEntry.Entry); err != nil {
				return err
			}
		} else if tEntry.Entry.GetType() == "tree" {
			if err := WriteTree(tEntry.Entry.(Tree)); err != nil {
				return err
			}
		}
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

var EmptyTreeError = errors.New("cannot create an empty tree")

// Create tree object for a given directory.
func CreateTreeObject(dirpath string) (Tree, error) {
	if fi, err := os.Stat(dirpath); err != nil || !fi.IsDir() {
		return Tree{}, errors.New("cannot create tree from a file")
	}
	dirEntries, err := os.ReadDir(dirpath)
	if err != nil {
		return Tree{}, err
	}
	// Directory without any entries is ignored.
	if len(dirEntries) == 0 {
		return Tree{}, EmptyTreeError
	}
	var t Tree
	for _, dirEntry := range dirEntries {
		tEntry := treeEntry{}
		dirEntryPath := filepath.Join(dirpath, dirEntry.Name())
		var (
			object Object
			mode   string
			err    error
		)
		if dirEntry.IsDir() {
			if dirEntry.Name() == GITDIR {
				continue
			}
			object, err = CreateTreeObject(dirEntryPath)
			// Once more: we skip empty trees.
			if err == EmptyTreeError {
				continue
			} else if err != nil {
				return Tree{}, err
			}
			mode = "040000" // directory aka tree
		} else {
			object, err = CreateBlobObject(dirEntryPath)
			if err != nil {
				return Tree{}, err
			}
			// TODO: Handle all permission bits properly.
			// 100755 - executable
			// 120000 - symlink
			mode = "100644" // normal file
		}
		hash, err := GetHash(object)
		if err != nil {
			return Tree{}, err
		}
		tEntry.Mode = mode
		tEntry.Hash = hash
		tEntry.Name = dirEntry.Name()
		tEntry.Entry = object
		t = append(t, tEntry)
	}
	return t, nil
}

func CreateCommitObject(treehash string, message string) (Commit, error) {
	author, err := getAuthorFromConfig()
	if err != nil {
		return Commit{}, err
	}
	headCommitHash, err := getHeadCommitHash()
	if err != nil {
		return Commit{}, err
	}
	return Commit{
		TreeHash:   treehash,
		ParentHash: headCommitHash,
		Author:     author,
		Time:       time.Now(),
		Msg:        message,
	}, nil
}

func getAuthorFromConfig() (AuthorType, error) {
	return AuthorType{
		Name:  "Antoni Szczepanik",
		Email: "szczepanik.antoni@gmail.com",
	}, nil
}

func exists(hash string) error {
	objectDir, err := GetGitSubdir("objects")
	if err != nil {
		return err
	}
	objectSubDir, objectName, err := splitHash(hash)
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

type EmptyObjectError struct {
	Msg     string
	Content []byte
	Type    string
}

func (o EmptyObjectError) Error() string {
	return fmt.Sprintf("%s\n%s\n%s\n", o.Msg, o.Type, string(o.Content))
}

// Check if object was properly initialized.
func isEmpty(o Object) error {
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

// Get position of the null byte, to seperate header from content.
func getNullBytePos(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
