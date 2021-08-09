package objects

import (
	"fmt"
	"time"

	"github.com/antoniszczepanik/gggit/refs"
)

const commit string = "commit"

type Commit struct {
	TreeHash   string
	ParentHash string
	Author     AuthorType
	// TODO: Should be represented as in RFC 2822. On drive it will be stored
	// as unixts timezoneoffset
	Time time.Time
	Msg  string
}

type AuthorType struct {
	Name  string
	Email string
}

func (c Commit) GetContent() ([]byte, error) {
	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	if c.ParentHash != "" {
		content += fmt.Sprintf("parent %s\n", c.ParentHash)
	}
	_, offset := c.Time.Zone()
	time := fmt.Sprintf("%d %d", c.Time.Unix(), offset)
	content += fmt.Sprintf("author %s <%s> %s\n\n", c.Author.Name, c.Author.Email, time)
	content += fmt.Sprintf("%s\n", c.Msg)
	return []byte(content), nil
}

func (c Commit) GetType() string {
	return commit
}

func parseCommit(contents []byte) (Commit, error) {
	return Commit{}, nil
}

func CreateCommitObject(treehash string, message string) (Commit, error) {
	author, err := getAuthorFromConfig()
	if err != nil {
		return Commit{}, err
	}
	headCommitHash, err := refs.GetHeadCommitHash()
	if err != nil {
		return Commit{}, err
	}
	return Commit{
		TreeHash:   treehash,
		ParentHash: headCommitHash,
		Author:     author,
		Time:       time.Now(),
		Msg:        message,
	}, nil
}

func getAuthorFromConfig() (AuthorType, error) {
	return AuthorType{
		Name:  "Antoni Szczepanik",
		Email: "szczepanik.antoni@gmail.com",
	}, nil
}
