package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateRepository(t *testing.T) {
	cwd, _ := os.Getwd()
	test_directory := filepath.Join(cwd, "testdir")
	os.Mkdir(test_directory, 0755)
	defer os.RemoveAll(test_directory)


	type testCase struct {
		path  string
		isDir bool
	}

	testCases := []testCase{
		{path: ".gggit", isDir: true},
		{path: ".gggit/objects", isDir: true},
		{path: ".gggit/branches", isDir: true},
		{path: ".gggit/refs/tags", isDir: true},
		{path: ".gggit/refs/heads", isDir: true},
		{path: ".gggit/HEAD", isDir: false},
		{path: ".gggit/description", isDir: false},
	}

	repoPath, err := createRepository(test_directory)
	if err != nil {
		t.Errorf("error creating the repository: %v\n", err)
		return
	}

	for _, testCase := range testCases {
		fullPath := filepath.Join(repoPath, testCase.path)
		file_info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			t.Error(fmt.Sprintf("%v does not exist", fullPath))
			return
		}
		if file_info.IsDir() != testCase.isDir {
			t.Error(fmt.Sprintf("%v is of wrong type", fullPath))
			return
		}
	}
}
