package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/objects"
	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/utils"
)

func Commit(args []string) {
	// TODO: Split this to multiple methods. Should cmd helpers be in main?
	repoRoot, err := utils.GetRepoRoot("")
	if err != nil {
		utils.Usage("not a git repository (or any of the parent directories)")
	}
	hash, err := objects.HashTree(repoRoot, true)
	if err != nil {
		fmt.Println(err)
		utils.Usage("failed to hash current tree")
	}
	// TODO: Add possibility to specify own message.
	msg := "Hello from gggit."
	c, err := objects.CreateCommitObject(hash, msg)
	if err != nil {
		utils.Usage("failed to create commit object")
	}
	err = objects.Write(c)
	if err != nil {
		utils.Usage("failed to write a commit object")
	}
	commitHash, err := objects.GetHash(c)
	if err != nil {
		utils.Usage("could not get hash for new commit")
	}
	refPath, err := refs.GetCurrentRef()
	if err != nil {
		utils.Usage("cannot get current ref. Are you in detached HEAD mode?")
	}
	err = refs.UpdateRef(refPath, commitHash)
	if err != nil {
		fmt.Println(err)
		utils.Usage("cannot update current ref")
	}
	err = refs.CheckoutRef(refPath)
	if err != nil {
		fmt.Println(err)
		utils.Usage("could not checkout the new ref")
	}
	fmt.Println("commit " + commitHash)
	err = objects.PrintObject(commitHash)
	if err != nil {
		utils.Usage("could not print commit content")
	}
}
