package refs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antoniszczepanik/gggit/utils"
)

var ErrMissingRef = errors.New("head points to a missing ref")

type HeadPointer struct {
	content string
}

// Resolve hash of commit HEAD points at.
func (hp HeadPointer) hash() (string, error) {
	isDetached, err := hp.detached()
	if err != nil {
		return "", err
	}
	// Content includes a line feed.
	if isDetached {
		return removeLastChar(hp.content), nil
	}
	refPath, err := parseRef(hp.content)
	if err != nil {
		return "", err
	}
	return ReadHashFromRef(refPath)
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
	f, err := utils.GetGitFile("HEAD")
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

// Get path of the ref that HEAD is currently pointing at.
func GetCurrentRefPath() (string, error) {
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

func GetHeadCommitHash() (string, error) {
	head, err := readHeadPointer()
	if err != nil {
		return "", err
	}
	return head.hash()
}

// Returns empty string if ref does not exist yet.
func ReadHashFromRef(refPath string) (string, error) {
	f, err := utils.GetGitFile(refPath)
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
func CreateNewRef(name string) (*os.File, error) {
	headsDir, err := utils.GetGitSubdir("refs/heads")
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

// Creates ref file if does not exists.
func PointRefAt(refPath string, commitHash string) error {
	filePath, err := utils.GetGitFilePath(refPath)
	if err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("%s\n", commitHash)))
	if err != nil {
		return err
	}
	return nil
}

func PointHeadAt(refPath string) error {
	headPath, err := utils.GetGitFilePath("HEAD")
	if err != nil {
		return err
	}
	f, err := os.Create(headPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("ref: %s\n", refPath)))
	if err != nil {
		return err
	}
	return nil
}

func GetRefPath(branchName string) string {
	return "refs/heads/" + branchName
}

func Exists(refName string) bool {
	refPath := GetRefPath(refName)
	refAbsPath, err := utils.GetGitFilePath(refPath)
	if err != nil {
		return false
	}
	if _, err := os.Stat(refAbsPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Remove line feeds and whatnot.
func removeLastChar(text string) string {
	if len(text) == 0 {
		return ""
	}
	return text[:len(text)-2]
}
