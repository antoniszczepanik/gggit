package objects

import (
	"fmt"
	"time"
)

const CommitObject ObjectType = "commit"

type Commit struct {
	TreeHash   string
	ParentHash string
	Author     Author
	// TODO: Should be represented as in RFC 2822. On drive it will be stored
	// as unixts timezoneoffset
	Time time.Time
	Msg  string
}

type Author struct {
	Name  string
	Email string
}

func (c Commit) GetContent() (string, error) {
	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	if c.ParentHash != "" {
		content += fmt.Sprintf("parent %s\n", c.ParentHash)
	}
	_, offset := c.Time.Zone()
	time := fmt.Sprintf("%d %d", c.Time.Unix(), offset)
	content += fmt.Sprintf("author %s <%s> %s\n\n", c.Author.Name, c.Author.Email, time)
	content += fmt.Sprintf("%s\n", c.Msg)
	return content, nil
}

func (c Commit) GetType() ObjectType {
	return CommitObject
}

func (c Commit) Write() error {
	return Write(c)
}

func ReadCommit(hash string) (Commit, error) {
	rawContent, err := getObjectRawContent(hash)
	if err != nil {
		return Commit{}, err
	}
	_, _, _, err = splitRawContent(rawContent)
	if err != nil {
		return Commit{}, err
	}
	return Commit{
		TreeHash: "this is a tree hash of dummy commit",
		ParentHash: "",
		Author: Author{
			Name: "Antoni Szczepanik",
			Email: "szczepanik.antoni@gmail.com",
		},
		Time: time.Now(),
		Msg: "dummy commit message",
	}, nil
}

func parseCommit(contents string) (Commit, error) {
	// TODO: Add method to parse contents of commit from file.
	contents += "a"
	return Commit{}, nil
}

func CreateCommitObject(treeHash string, parentHash string, message string) (Commit, error) {
	author, err := getAuthorFromConfig()
	if err != nil {
		return Commit{}, err
	}
	if err != nil {
		return Commit{}, err
	}
	return Commit{
		TreeHash:   treeHash,
		ParentHash: parentHash,
		Author:     author,
		Time:       time.Now(),
		Msg:        message,
	}, nil
}

func getAuthorFromConfig() (Author, error) {
	return Author{
		Name:  "Antoni Szczepanik",
		Email: "szczepanik.antoni@gmail.com",
	}, nil
}
