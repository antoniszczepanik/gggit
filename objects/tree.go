package objects

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/antoniszczepanik/gggit/utils"
)

const tree string = "tree"

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
		entryContent := fmt.Sprintf(treeEntryFmt+"\n", e.Mode, e.Entry.GetType(), e.Hash, e.Name)
		content = append(content, []byte(entryContent)...)
	}
	return content, nil
}

func (t Tree) GetType() string {
	return tree
}

func parseTree(contents []byte) (Tree, error) {
	var (
		t     Tree
		emode string
		etype string
		ehash string
		ename string
	)
	rawEntries := utils.SplitEntries(contents, '\n')
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

// Write tree object, recursively writing all its entries.
func WriteTree(t Tree) error {
	if err := Write(t); err != nil {
		return err
	}
	for _, tEntry := range t {
		if Exists(tEntry.Hash) == nil {
			continue
		}
		if tEntry.Entry.GetType() == blob {
			if err := Write(tEntry.Entry); err != nil {
				return err
			}
		} else if tEntry.Entry.GetType() == tree {
			if err := WriteTree(tEntry.Entry.(Tree)); err != nil {
				return err
			}
		}
	}
	return nil
}

var ErrEmptyTree = errors.New("cannot create an empty tree")

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
		return Tree{}, ErrEmptyTree
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
			if dirEntry.Name() == utils.GITDIR {
				continue
			}
			object, err = CreateTreeObject(dirEntryPath)
			// Once more: we skip empty trees.
			if errors.Is(err, ErrEmptyTree) {
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

// Assumes caller verified that path points at a directory.
func HashTree(path string, write bool) (string, error) {
	t, err := CreateTreeObject(path)
	if errors.Is(err, ErrEmptyTree) {
		return "", errors.New("directory is empty")
	} else if err != nil {
		return "", err
	}
	if write {
		if err := WriteTree(t); err != nil {
			return "", err
		}
	}
	return GetHash(t)
}
