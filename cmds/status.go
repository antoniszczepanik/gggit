package cmds

import (
	"os"
	"strings"
	"fmt"

	"github.com/antoniszczepanik/gggit/refs"
	"github.com/antoniszczepanik/gggit/utils"
)

func Status(args []string) {
	currentCommitHash, err := refs.GetHeadCommitHash()
	if err != nil {
		utils.Usage("could not get current commit")
	}
	refPath, err := refs.GetCurrentRefPath()
	if err != nil {
		fmt.Printf("detached HEAD mode on %s\n", currentCommitHash)
		return
	}
	fmt.Printf("On branch %s (commit %s)\n", getBranchName(refPath), currentCommitHash)
}

func getBranchName(refPath string) string {
	words := strings.Split(refPath, string(os.PathSeparator))
	if len(words) < 1 {
		return ""
	}
	return words[len(words)-1]
}
