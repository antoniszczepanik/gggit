package cmds

import (
	"os"
	"strings"

	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/utils"
)

func Status(args []string) {
	currentCommitHash, err := refs.GetHeadCommitHash()
	if err != nil {
		utils.Usage("could not get current commit")
	}
	refPath, err := refs.GetCurrentRef()
	if err != nil {
		utils.Usage("detached HEAD mode on " + currentCommitHash)
	}
	utils.Usage("On branch " + getBranchName(refPath))
}

func getBranchName(refPath string) string {
	words := strings.Split(refPath, string(os.PathSeparator))
	if len(words) < 1 {
		return ""
	}
	return words[len(words)-1]
}
