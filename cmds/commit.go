package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/objects"
	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/common"
)

func Commit(args []string) {
	repoRoot, err := common.GetRepoRoot("")
	if err != nil {
		common.Usage("not a git repository (or any of the parent directories)")
	}

	treeHash, err := objects.HashTree(repoRoot, true)
	if err != nil {
		common.Usage(err.Error())
	}
	// TODO: Add possibility to specify own message.
	msg := "Hello from gggit."

	parentHash, err := refs.GetHeadCommitHash()
	if err == refs.ErrBranchWithoutHash {
		parentHash = ""
	} else if err != nil {
		common.Usage(err.Error())
	}
	c, err := objects.CreateCommitObject(treeHash, parentHash, msg)
	if err != nil {
		common.Usage("failed to create commit object")
	}
	err = objects.Write(c)
	if err != nil {
		common.Usage("failed to write a commit object")
	}
	commitHash, err := objects.CalculateHash(c)
	if err != nil {
		common.Usage("could not get hash for new commit")
	}
	branchName, err := refs.GetCurrentBranch()
	if err != nil {
		common.Usage("cannot get current ref. Are you in detached HEAD mode?")
	}
	err = refs.PointBranchAt(branchName, commitHash)
	if err != nil {
		fmt.Println(err)
		common.Usage("cannot update current ref")
	}
	err = refs.PointHeadAtBranch(branchName)
	if err != nil {
		fmt.Println(err)
		common.Usage("could not checkout the new ref")
	}
	fmt.Printf("commit %s\n", commitHash)
	err = objects.PrintObject(commitHash)
	if err != nil {
		common.Usage("could not print commit content")
	}
}
