package objects

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (t Tree) GetContent() (string, error) {
	content := ""
	for _, e := range t {
		if e.Mode == "" || e.Hash == "" || e.Name == "" || e.Entry == nil {
			return "", errors.New("cannot get content of tree with missing attributes")
		}
		entryContent := fmt.Sprintf(treeEntryFmt+"\n", e.Mode, e.Entry.GetType(), e.Hash, e.Name)
		content += entryContent
	}
	return content, nil
}

func (t Tree) GetType() string {
	return tree
}

func (t Tree) Write() error {
	if err := Write(t); err != nil {
		return err
	}
	// For a tree recursively write all of its contents.
	for _, tEntry := range t {
		// No need to write existing tree objects.
		if Exists(tEntry.Hash) == nil {
			continue
		}
		if err := tEntry.Entry.Write(); err != nil {
			return err
		}
	}
	return nil
}

// Write out tree entries to working directory aka. checokut.
// atPath is a path of a directory tree objects should be written to.
// func (t Tree) UpdateWorkdir(atPath string) error {
//	for _, tEntry := range t {
//		entry := tEntry.Entry
//		entryType := entry.GetType()
//		if entryType == blob {
//			_, err := entry.GetContent()
//			if err != nil {
//				return err
//			}
//			fmt.Printf("Would write contents to %s\n", atPath + tEntry.Name)
//		} else if entryType == tree {
//			if err := entry.UpdateWorkdir(atPath+tEntry.Name); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

func parseTree(contents string) (Tree, error) {
	var (
		t         Tree
		entryMode string
		entryType string
		entryHash string
		entryName string
	)
	rawEntries := strings.Split(contents, "\n")
	for _, rawEntry := range rawEntries {
		r := strings.NewReader(rawEntry)
		_, err := fmt.Fscanf(r, treeEntryFmt, &entryMode, &entryType, &entryHash, &entryName)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Tree{}, err
		}
		entryObj, err := Read(entryHash)
		if err != nil {
			return Tree{}, err
		}
		t = append(t, treeEntry{Mode: entryMode, Hash: entryHash, Name: entryName, Entry: entryObj})
	}
	return t, nil
}

var ErrEmptyTree = errors.New("cannot create an empty tree")

func NewTreeFromDirectory(dirpath string) (Tree, error) {
	if fi, err := os.Stat(dirpath); err != nil || !fi.IsDir() {
		return Tree{}, errors.New("cannot create tree from a file")
	}
	dirEntries, err := os.ReadDir(dirpath)
	if err != nil {
		return Tree{}, err
	}
	// Ignore directories without any entries.
	if len(dirEntries) == 0 {
		return Tree{}, ErrEmptyTree
	}
	var t Tree
	for _, dirEntry := range dirEntries {
		dirEntryPath := filepath.Join(dirpath, dirEntry.Name())
		var (
			object Object
			mode   string
			err    error
		)
		// TODO: I hate this if. Refactor it, please!
		if dirEntry.IsDir() {
			// TODO: Should be handled by .gitignore not hardcoded.
			if dirEntry.Name() == utils.GitDirName {
				continue
			}
			object, err = NewTreeFromDirectory(dirEntryPath)
			// Once more: we skip empty trees.
			if errors.Is(err, ErrEmptyTree) {
				continue
			} else if err != nil {
				return Tree{}, err
			}
			mode = "040000" // directory aka tree
		} else {
			object, err = NewBlobFromFile(dirEntryPath)
			if err != nil {
				return Tree{}, err
			}
			// TODO: Handle all permission bits properly.
			// 100755 - executable
			// 120000 - symlink
			mode = "100644" // normal file
		}
		hash, err := CalculateHash(object)
		if err != nil {
			return Tree{}, err
		}
		t = append(t, treeEntry{
			Mode:  mode,
			Hash:  hash,
			Name:  dirEntry.Name(),
			Entry: object,
		})
	}
	return t, nil
}

// Assumes caller verified that path points at a directory.
func HashTree(path string, write bool) (string, error) {
	t, err := NewTreeFromDirectory(path)
	if errors.Is(err, ErrEmptyTree) {
		return "", errors.New("directory is empty")
	} else if err != nil {
		return "", err
	}
	if write {
		if err := t.Write(); err != nil {
			return "", err
		}
	}
	return CalculateHash(t)
}
