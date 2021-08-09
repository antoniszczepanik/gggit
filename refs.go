package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var MissingRefError = errors.New("head points to a missing ref")

type HeadPointer struct {
	content string
}

func (hp HeadPointer) hash() (string, error) {
	isDetached, err := hp.detached()
	if err != nil {
		return "", err
	}
	if isDetached {
		// Content includes a line feed.
		return removeLastChar(hp.content), nil
	}
	refPath, err := parseRef(hp.content)
	if err != nil {
		return "", err
	}
	return readHashFromRef(refPath)
}

// Get ref path from contents of HEAD file.
func parseRef(headContent string) (string, error) {
	var refPath string
	r := strings.NewReader(headContent)
	_, err := fmt.Fscanf(r, "ref: %s\n", &refPath)
	if err != nil {
		return "", err
	}
	return refPath, nil
}

func (hp HeadPointer) detached() (bool, error) {
	if len(hp.content) < len("ref: refs/heads/a\n") {
		return false, errors.New("head points to invalid ref")
	}
	if hp.content[:5] == "ref: " {
		return false, nil
	}
	return true, nil
}

func readHeadPointer() (HeadPointer, error) {
	f, err := getGitFile("HEAD")
	if err != nil {
		return HeadPointer{}, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return HeadPointer{}, err
	}
	return HeadPointer{content: string(content)}, nil
}

func getCurrentRef() (string, error) {
	hp, err := readHeadPointer()
	if err != nil {
		return "", err
	}
	isDetached, err := hp.detached()
	if err != nil {
		return "", err
	}
	if isDetached {
		return "", errors.New("head is detached, could not get current ref")
	}
	refPath, err := parseRef(hp.content)
	if err != nil {
		return "", err
	}
	return refPath, nil
}

func getHeadCommitHash() (string, error) {
	head, err := readHeadPointer()
	if err != nil {
		return "", err
	}
	return head.hash()
}

// Returns empty string if ref does not exist yet.
func readHashFromRef(refPath string) (string, error) {
	f, err := getGitFile(refPath)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return removeLastChar(string(content)), nil
}

// Create new ref and return it's pointer. Caller is responsible for closing
// the file.
func createNewRef(name string) (*os.File, error) {
	headsDir, err := GetGitSubdir("refs/heads")
	if err != nil {
		return nil, err
	}
	newRefPath := filepath.Join(headsDir, name)
	newRefFile, err := os.Create(newRefPath)
	if err != nil {
		return nil, err
	}
	return newRefFile, nil
}

// Update ref contents to point at a given hash.
// Creates ref file if such does not exists.
func updateRef(refPath string, commitHash string) error {
	filePath, err := getGitFilePath(refPath)
	if err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("%s\n", commitHash)))
	if err != nil {
		return err
	}
	return nil
}

// Point HEAD at ref.
func checkoutRef(refPath string) error {
	headPath, err := getGitFilePath("HEAD")
	if err != nil {
		return err
	}
	f, err := os.Create(headPath)
	if err != nil {
		return nil
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("ref: %s\n", refPath)))
	if err != nil {
		return err
	}
	return nil
}

func getRefPath(branchName string) string {
	return "refs/heads/" + branchName
}

// Remove line feeds and whatnot.
func removeLastChar(text string) string {
	if len(text) == 0 {
		return ""
	}
	return text[:len(text)-2]
}
