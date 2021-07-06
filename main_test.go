package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateRepository(t *testing.T) {
	cwd, _ := os.Getwd()
	testDir := filepath.Join(cwd, "testdir")
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)

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

	repoPath, err := createRepository(testDir)
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

func TestHashObject(t *testing.T) {
	cwd, _ := os.Getwd()

	testDir := filepath.Join(cwd, "testdir")
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)
	os.Chdir(testDir)
	_, err := createRepository(testDir)
	if err != nil {
		t.Errorf("error creating the repository: %v\n", err)
		return
	}

	test_file := filepath.Join(testDir, "test_file.txt")
	f, err := os.Create(test_file)
	if err != nil {
		t.Error("hash-object fail: cannot create object to hash")
	}
	defer f.Close()

	_, err2 := f.WriteString("Ptaki latają kluczem\n")

	if err2 != nil {
		t.Error("hash-object fail: cannot write to object to hash")
	}
	// TODO:
	// Hash object.
	// Check if it was created with the proper name in proper subdir.
	// Cat object and verify if contents match.
}
