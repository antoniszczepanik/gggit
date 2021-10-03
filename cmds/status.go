package cmds

import (
	"os"
	"strings"
	"fmt"

	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/common"
)

func Status(args []string) {
	currentCommitHash, err := refs.GetHeadCommitHash()
	if err != nil {
		common.Usage("could not get current commit")
	}
	branchName, err := refs.GetCurrentBranch()
	if err == refs.ErrDetachedHead {
		fmt.Printf("detached HEAD mode on %s\n", currentCommitHash)
		return
	} else if err != nil {
		common.Usage("could not get current branch")
	}
	fmt.Printf("On branch %s (commit %s)\n", branchName, currentCommitHash)
}

func getBranchName(refPath string) string {
	words := strings.Split(refPath, string(os.PathSeparator))
	if len(words) < 1 {
		return ""
	}
	return words[len(words)-1]
}
