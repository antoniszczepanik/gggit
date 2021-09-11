package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/objects"
	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/utils"
)

func Commit(args []string) {
	repoRoot, err := utils.GetRepoRoot("")
	if err != nil {
		utils.Usage("not a git repository (or any of the parent directories)")
	}
	treeHash, err := objects.HashTree(repoRoot, true)
	if err != nil {
		utils.Usage(err.Error())
	}
	// TODO: Add possibility to specify own message.
	msg := "Hello from gggit."
	parentHash, err := refs.GetHeadCommitHash()
	if err != nil {
		utils.Usage(err.Error())
	}
	c, err := objects.CreateCommitObject(treeHash, parentHash, msg)
	if err != nil {
		utils.Usage("failed to create commit object")
	}
	err = objects.Write(c)
	if err != nil {
		utils.Usage("failed to write a commit object")
	}
	commitHash, err := objects.CalculateHash(c)
	if err != nil {
		utils.Usage("could not get hash for new commit")
	}
	refPath, err := refs.GetCurrentRefPath()
	if err != nil {
		utils.Usage("cannot get current ref. Are you in detached HEAD mode?")
	}
	err = refs.PointRefAt(refPath, commitHash)
	if err != nil {
		fmt.Println(err)
		utils.Usage("cannot update current ref")
	}
	err = refs.PointHeadAt(refPath)
	if err != nil {
		fmt.Println(err)
		utils.Usage("could not checkout the new ref")
	}
	fmt.Printf("commit %s\n", commitHash)
	err = objects.PrintObject(commitHash)
	if err != nil {
		utils.Usage("could not print commit content")
	}
}
