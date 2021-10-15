package cmds

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/antoniszczepanik/gggit/common"
)

func Init(args []string) {
	path, err := initRepository("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created new repository at %v\n", path)
}
func initRepository(path string) (string, error) {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("specified directory does not exist")
	}
	gitdir := filepath.Join(path, common.GitDirName)
	if _, err := os.Stat(gitdir); !os.IsNotExist(err) {
		return "", fmt.Errorf("git directory already exists at %v", path)
	}
	err = os.Mkdir(gitdir, 0755)
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(gitdir, "objects"), 0755)
	if err != nil {
		return "", err
	}
	err = os.Mkdir(filepath.Join(gitdir, "branches"), 0755)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(filepath.Join(gitdir, "refs", "tags"), 0755)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(filepath.Join(gitdir, "refs", "heads"), 0755)
	if err != nil {
		return "", err
	}
	headFile, err := os.Create(filepath.Join(gitdir, "HEAD"))
	if err != nil {
		return "", err
	}
	_, err = headFile.WriteString("ref: refs/heads/master")
	if err != nil {
		return "", err
	}
	descf, err := os.Create(filepath.Join(gitdir, "description"))
	if err != nil {
		return "", err
	}
	_, err = descf.WriteString("Unnamed repository; edit this file 'description' to name the repository.\n")
	if err != nil {
		return "", err
	}
	return path, nil
}
