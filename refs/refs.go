package refs

import (
	"errors"
	"regexp"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/antoniszczepanik/gggit/common"
	"github.com/antoniszczepanik/gggit/objects"
)

var ErrMissingRef = errors.New("HEAD points to a missing ref")

var ErrDetachedHead = errors.New("HEAD is in detached mode")

var refRegex = regexp.MustCompile(
	`ref: refs/(?P<type>heads|tags|remotes)/(?P<name>[a-zA-Z1-9-_]+$)`,
)

type headPointer struct {
	content string
}

// Resolve hash of commit HEAD points at.
func (hp headPointer) hash() (string, error) {
	isDetached, err := hp.detached()
	if err != nil {
		return "", err
	}
	// Content includes a line feed.
	if isDetached {
		return removeLastChar(hp.content), nil
	}
	_, branchName, err := parseRef(hp.content)
	if err != nil {
		return "", err
	}
	return ReadBranchHash(branchName)
}

func (hp headPointer) detached() (bool, error) {
	if len(hp.content) < len("ref: refs/heads/a\n") {
		return false, errors.New("head points to invalid ref")
	}
	if hp.content[:5] == "ref: " {
		return false, nil
	}
	return true, nil
}

// Get path of the ref that HEAD is currently pointing at.
func GetCurrentBranch() (string, error) {
	hp, err := readHeadPointer()
	if err != nil {
		return "", err
	}
	isDetached, err := hp.detached()
	if err != nil {
		return "", fmt.Errorf("check if detached: %w", err)
	}
	if isDetached {
		return "", ErrDetachedHead
	}
	_, branchName, err := parseRef(hp.content)
	if err != nil {
		return "", fmt.Errorf("parse ref: %w", err)
	}
	return branchName, nil
}

func readHeadPointer() (headPointer, error) {
	f, err := common.GetGitFile("HEAD")
	if err != nil {
		return headPointer{}, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return headPointer{}, err
	}
	return headPointer{content: string(content)}, nil
}

// Parse from contents of HEAD file into ref type and ref name.
func parseRef(headContent string) (string, string, error) {
	match := refRegex.FindStringSubmatch(headContent)
	if match == nil || len(match) != 3 {
		fmt.Printf("Cannot parse head content: %s\n", headContent)
		return "", "", fmt.Errorf("parse head content: %s", headContent)
	}
	return match[1], match[2], nil
}



func GetHeadTreeHash() (string, error){
	commitHash, err := GetHeadCommitHash()
	if err != nil {
		return "", err
	}
	commit, err := objects.ReadCommit(commitHash)
	if err != nil {
		return "", err
	}
	return commit.TreeHash, nil
}

func GetHeadCommitHash() (string, error) {
	head, err := readHeadPointer()
	if err != nil {
		return "", err
	}
	return head.hash()
}

var ErrBranchWithoutHash = errors.New("branch does not have any commits yet")

// Returns empty string if ref does not exist yet.
func ReadBranchHash(branchName string) (string, error) {
	branchRefPath := getRefPath(branchName)
	ref, err := common.GetGitFile(branchRefPath)
	if err != nil{
		return "", ErrBranchWithoutHash
	}
	defer ref.Close()

	content, err := io.ReadAll(ref)
	if err != nil {
		return "", err
	}
	return removeLastChar(string(content)), nil
}

// Create new ref and return it's pointer. Caller is responsible for closing
// the file.
func CreateNewRef(name string) (*os.File, error) {
	headsDir, err := common.GetGitSubdir("refs/heads")
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

// Point branch pointer at commit.
func PointBranchAt(branchName, commitHash string) error {
	branchRefPath := getRefPath(branchName)
	branchRefPathAbs, err := common.GetGitFilePath(branchRefPath)
	if err != nil {
		return fmt.Errorf("point branch at commit: %w", err)
	}

	f, err := os.Create(branchRefPathAbs)
	if err != nil {
		return fmt.Errorf("overwrite branch pointer file: %w", err)
	}
	defer f.Close()

	_, err = f.Write([]byte(fmt.Sprintf("%s\n", commitHash)))
	if err != nil {
		return err
	}

	return nil
}

func PointHeadAtBranch(branchName string) error {
	headPath, err := common.GetGitFilePath("HEAD")
	if err != nil {
		return err
	}
	f, err := os.Create(headPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("ref: %s", getRefPath(branchName))))
	if err != nil {
		return err
	}
	return nil
}

func Exists(branchName string) bool {
	branchRefPath := getRefPath(branchName)
	branchRefPathAbs, err := common.GetGitFilePath(branchRefPath)
	if err != nil {
		return false
	}
	if _, err := os.Stat(branchRefPathAbs); os.IsNotExist(err) {
		return false
	}
	return true
}

func getRefPath(branchName string) string {
	return "refs/heads/" + branchName
}

// Remove line feeds and whatnot.
func removeLastChar(text string) string {
	if len(text) == 0 {
		return ""
	}
	return text[:len(text)-1]
}
