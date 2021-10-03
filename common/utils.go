package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const GitDirName = ".gggit"

// Find git directory and return its specific subdirectory.
func GetGitSubdir(subdirName string) (string, error) {
	gitDir, err := GetGitDir("")
	if err != nil {
		return "", err
	}
	subDir := filepath.Join(gitDir, subdirName)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory %s does not exist", subDir)
	}
	return subDir, nil
}

// Get a path to .git directory.
func GetGitDir(path string) (string, error) {
	repoDir, err := GetRepoRoot(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoDir, GitDirName), nil
}

// Get a path to repository root.
func GetRepoRoot(path string) (string, error) {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	gitPath := filepath.Join(path, GitDirName)
	if _, err = os.ReadDir(gitPath); os.IsNotExist(err) {
		if path == "/" {
			return "", errors.New("did not find git directory")
		}
		return GetRepoRoot(filepath.Dir(path))
	}
	return path, nil
}

// Returns a pointer to internal git file. Caller is responsilbe for
// closing a file handle.
func GetGitFile(filename string) (*os.File, error) {
	fullPath, err := GetGitFilePath(filename)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Get full path to a git internal file.
// Accepts filename relative to .git directory.
func GetGitFilePath(filename string) (string, error) {
	gitDir, err := GetGitDir("")
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, filename), nil
}

// Split hash to get directory and filename, so that
// serialized objects are scattered among directories.
func SplitHash(hash string) (string, string, error) {
	if len(hash) != 40 {
		return "", "", errors.New("incorrect hash length")
	}
	return hash[:2], hash[2:], nil
}


func Usage(msg string) {
	_, err := io.WriteString(os.Stderr, msg+"\n")
	if err != nil {
		panic("I have no idea what's going on. Do you?")
	}
	os.Exit(1)
}
