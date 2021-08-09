package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	case "status":
		CmdStatus(args)
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
	// TODO: Split this to multiple methods. Should cmd helpers be in main?
	repoRoot, err := getRepoRoot("")
	if err != nil {
		usage("not a git repository (or any of the parent directories)")
	}
	hash, err := hashTree(repoRoot, true)
	if err != nil {
		fmt.Println(err)
		usage("failed to hash current tree")
	}
	// TODO: Add possibility to specify own message.
	msg := "Hello from gggit."
	c, err := CreateCommitObject(hash, msg)
	if err != nil {
		usage("failed to create commit object")
	}
	err = Write(c)
	if err != nil {
		usage("failed to write a commit object")
	}
	commitHash, err := GetHash(c)
	if err != nil {
		usage("could not get hash for new commit")
	}
	refPath, err := getCurrentRef()
	if err != nil {
		usage("cannot get current ref. Are you in detached HEAD mode?")
	}
	err = updateRef(refPath, commitHash)
	if err != nil {
		fmt.Println(err)
		usage("cannot update current ref")
	}
	err = checkoutRef(refPath)
	if err != nil {
		fmt.Println(err)
		usage("could not checkout the new ref")
	}
	fmt.Println("commit " + commitHash)
	err = PrintObject(commitHash)
	if err != nil {
		usage("could not print commit content")
	}
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
			usage(fmt.Sprintf("%s is not a valid option", args[0]))
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
	if err == EmptyTreeError {
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

func CmdStatus(args []string) {
	currentCommitHash, err := getHeadCommitHash()
	if err != nil {
		usage("could not get current commit")
	}
	refPath, err := getCurrentRef()
	if err != nil {
		usage("detached HEAD mode on " + currentCommitHash)
	}
	usage("On branch " + getBranchName(refPath))
}

func getBranchName(refPath string) string {
	words := strings.Split(refPath, string(os.PathSeparator))
	if len(words) < 1 {
		return ""
	}
	return words[len(words)-1]
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
		return "", fmt.Errorf("git directory already exists at %v\n", path)
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
	headf, err := os.Create(filepath.Join(gitdir, "HEAD"))
	if err != nil {
		return "", err
	}
	_, err = headf.WriteString("ref: refs/heads/master\n")
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

// Find git directory and return its specific subdirectory.
func GetGitSubdir(subdirName string) (string, error) {
	gitDir, err := getGitDir("")
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
func getGitDir(path string) (string, error) {
	repoDir, err := getRepoRoot(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoDir, GITDIR), nil
}

// Get a path to repository root.
func getRepoRoot(path string) (string, error) {
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
			return getRepoRoot(filepath.Dir(path))
		}
	}
	return path, nil
}

// Returns a pointer to internal git file. Caller is responsilbe for
// closing a file handle.
func getGitFile(filename string) (*os.File, error) {
	fullPath, err := getGitFilePath(filename)
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
func getGitFilePath(filename string) (string, error) {
	gitDir, err := getGitDir("")
	if err != nil {
		return "", err
	}
	return filepath.Join(gitDir, filename), nil
}

func usage(msg string) {
	_, err := io.WriteString(os.Stderr, msg+"\n")
	if err != nil {
		panic("whut")
	}
	os.Exit(1)
}
