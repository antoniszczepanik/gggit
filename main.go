package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const GITDIR string = ".gggit"

func main() {
	if len(os.Args) < 2 {
		usage("You need to specify a gggit command.")
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "add":
		CmdAdd(args)
	case "cat-file":
		CmdCat(args)
	case "checkout":
		CmdCheckout(args)
	case "commit":
		CmdCommit(args)
	case "hash-object":
		CmdHash(args)
	case "init":
		CmdInit(args)
	case "log":
		CmdLog(args)
	case "ls-tree":
		CmdLs(args)
	case "ls-objects":
		CmdLsObjects(args)
	default:
		usage(fmt.Sprintf("Command %v is not available. Did you mean sth else?\n", cmd))
	}
}

func CmdAdd(args []string) {
	fmt.Println("add")
}

func CmdCat(args []string) {
	if len(args) != 1 {
		usage("You should provide hash of object to cat.")
	}
	err := PrintObject(args[0])
	if err != nil {
		fmt.Println(err)
	}
}

func CmdCheckout(args []string) {
	fmt.Println("checkout")
}

func CmdCommit(args []string) {
	fmt.Println("commit")
}

func CmdHash(args []string) {
	switch len(args) {
	case 0:
		usage("You should provide name of an entity to hash.")
	case 1:
		hash, err := hashEntityByType(args[0], false)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(hash)
	case 2:
		if args[0] != "-w" {
			usage(fmt.Sprintf("%s is not a valid option"))
			return
		}
		hash, err := hashEntityByType(args[1], true)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(hash)
	default:
		usage("Too many arguments")
	}
}

func hashEntityByType(path string, write bool) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if fileInfo.IsDir() {
		return hashTree(path, write)
	}
	return hashFile(path, write)
}

// Assumes caller verified that path points at a directory.
func hashTree(path string, write bool) (string, error) {
	t, err := CreateTreeObject(path)
	if err != nil {
		return "", err
	}
	if write {
		if err := WriteTree(t); err != nil {
			return "", err
		}
	}
	return GetHash(t)
}

func hashFile(path string, write bool) (string, error) {
	object, err := CreateBlobObject(path)
	if err != nil {
		return "", err
	}
	if write {
		if err := Write(object); err != nil {
			return "", err
		}
	}
	return GetHash(object)
}

func CmdInit(args []string) {
	path, err := initRepository("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created new repository at %v\n", path)
}

func CmdLs(args []string) {
	fmt.Println("ls-tree")
}

func CmdLog(args []string) {
	fmt.Println("log")
}

func CmdLsObjects(args []string) {
	objectsDir, err := GetGitSubdir("objects")
	if err != nil {
		usage("could not find git objects dir")
	}
	dirEntries, err := os.ReadDir(objectsDir)
	if err != nil {
		usage("could not read git objects dir")
	}
	for _, e := range dirEntries {
		subDirEntries, err := os.ReadDir(filepath.Join(objectsDir, e.Name()))
		if err != nil {
			usage("could not read one of object subdirs")
		}
		for _, se := range subDirEntries {
			fmt.Println(e.Name() + se.Name())
		}
	}

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
	gitdir := filepath.Join(path, GITDIR)
	if _, err := os.Stat(gitdir); !os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("git directory already exists at %v\n", path))
	}
	os.Mkdir(gitdir, 0755)
	os.Mkdir(filepath.Join(gitdir, "objects"), 0755)
	os.Mkdir(filepath.Join(gitdir, "branches"), 0755)
	os.MkdirAll(filepath.Join(gitdir, "refs", "tags"), 0755)
	os.MkdirAll(filepath.Join(gitdir, "refs", "heads"), 0755)
	headf, _ := os.Create(filepath.Join(gitdir, "HEAD"))
	headf.WriteString("ref: refs/heads/master\n")
	descf, _ := os.Create(filepath.Join(gitdir, "description"))
	descf.WriteString("Unnamed repository; edit this file 'description' to name the repository.\n")
	return path, nil
}

// Find git directory and return its specific subdirectory.
func GetGitSubdir(subdirName string) (string, error) {
	gitDir, err := getGitDir("")
	if err != nil {
		return "", err
	}
	subDir := filepath.Join(gitDir, subdirName)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		return "", errors.New(fmt.Sprintf("directory %s does not exist", subDir))
	}
	return subDir, nil
}

// Recursively try to find git directory in current or parent directory.
func getGitDir(path string) (string, error) {
	var err error
	if path == "" {
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	gitPath := filepath.Join(path, GITDIR)
	if _, err = os.ReadDir(gitPath); os.IsNotExist(err) {
		if path == "/" {
			return "", errors.New("did not find git directory")
		} else {
			return getGitDir(filepath.Dir(path))
		}
	}
	return gitPath, nil
}

func usage(msg string) {
	io.WriteString(os.Stderr, msg+"\n")
	os.Exit(1)
}
