package objects

import (
	"fmt"
	"strings"
	"time"
)

const CommitObject ObjectType = "commit"

type Commit struct {
	TreeHash   string
	ParentHash string
	Author     Author
	Time       time.Time
	Msg        string
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
	content += fmt.Sprintf(
		"author %s <%s> %s\n\n", c.Author.Name, c.Author.Email, c.Time.Format(time.RFC822Z))
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
	objectType, _, content, err := splitRawContent(rawContent)
	if err != nil {
		return Commit{}, err
	}
	if objectType != CommitObject {
		return Commit{}, fmt.Errorf("could not read commit %s: invalid object type '%s'", hash, objectType)
	}
	return parseCommit(content)
}

func parseCommit(content string) (Commit, error) {
	var (
		tree, parent, message string
		author                Author
		err                   error
		t                     time.Time
	)
	lines := strings.Split(content, "\n")
	isMessage := false
	for _, line := range lines {
		if line == "" {
			isMessage = true
			continue
		}
		if !isMessage {
			values := strings.SplitN(line, " ", 2)
			if len(values) == 2 {
				switch values[0] {
				case "tree":
					tree = values[1]
				case "parent":
					parent = values[1]
				case "author":
					author, t, err = parseAuthor(values[1])
					if err != nil {
						return Commit{}, err
					}
				}
			}
		} else {
			message += line
		}
	}
	return Commit{
		TreeHash:   tree,
		ParentHash: parent,
		Author:     author,
		Time:       t,
		Msg:        message,
	}, nil
}

func parseAuthor(value string) (Author, time.Time, error) {
	email_start := strings.Index(value, "<")
	email_end := strings.Index(value, ">")
	if email_start == -1 || email_end == -1 || email_end < email_start {
		return Author{}, time.Time{}, fmt.Errorf("did not find email delims (<>) in %s", value)
	}

	name := strings.Trim(value[:email_start], " ")
	email := value[email_start+1 : email_end]

	t, err := time.Parse(time.RFC822Z, value[email_end+2:])
	if err != nil {
		return Author{}, time.Time{}, fmt.Errorf("could not parse time: %w", err)
	}
	return Author{Name: name, Email: email}, t, nil
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
