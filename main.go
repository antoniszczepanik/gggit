package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const GITDIR string = ".gggit"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You need to specify a gggit command.")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "add":
		cmd_add(os.Args[2:])
	case "cat-file":
		cmd_cat(os.Args[2:])
	case "checkout":
		cmd_checkout(os.Args[2:])
	case "commit":
		cmd_commit(os.Args[2:])
	case "hash-object":
		cmd_hash(os.Args[2:])
	case "init":
		cmd_init(os.Args[2:])
	case "log":
		cmd_log(os.Args[2:])
	case "ls-tree":
		cmd_ls(os.Args[2:])
	case "merge":
		cmd_merge(os.Args[2:])
	case "rebase":
		cmd_rebase(os.Args[2:])
	case "rev-parse":
		cmd_rev(os.Args[2:])
	case "rm":
		cmd_rm(os.Args[2:])
	case "show-ref":
		cmd_show(os.Args[2:])
	case "tag":
		cmd_tag(os.Args[2:])
	default:
		fmt.Printf("Command %v is not available. Did you mean sth else?\n", cmd)
		os.Exit(1)
	}
}

func cmd_add(args []string) {
	fmt.Println("add")
}
func cmd_cat(args []string) {
	if len(args) != 1 {
		fmt.Println("You should provide hash of object to cat.")
		return
	}
	o, err := ReadObject(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	content, err := o.GetContent()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(string(content))
}
func cmd_checkout(args []string) {
	fmt.Println("checkout")
}
func cmd_commit(args []string) {
	fmt.Println("commit")
}
func cmd_hash(args []string) {
	if len(args) == 0 {
		fmt.Println("You should provide name of a file to hash.")
		return
	}
	var hash string
	var err error
	// Do not write by default.
	if len(args) == 1 {
		hash, err = hashFile(args[0], false)
	} else if args[0] == "-w" {
		hash, err = hashFile(args[1], true)
	} else {
		fmt.Println(args[0], "not supported. Did you mean sth else?")
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hash)
}
func cmd_init(args []string) {
	path, err := initRepository("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Created new repository at %v\n", path)
}
func cmd_log(args []string) {
	fmt.Println("log")
}
func cmd_ls(args []string) {
	fmt.Println("ls-tree")
}
func cmd_merge(args []string) {
	fmt.Println("merge")
}
func cmd_rebase(args []string) {
	fmt.Println("rebase")
}
func cmd_rev(args []string) {
	fmt.Println("rev-parse")
}
func cmd_rm(args []string) {
	fmt.Println("rm")
}
func cmd_show(args []string) {
	fmt.Println("show-ref")
}
func cmd_tag(args []string) {
	fmt.Println("tag")
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
