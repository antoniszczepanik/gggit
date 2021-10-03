package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/common"
	"github.com/antoniszczepanik/gggit/refs"
)

func Checkout(args []string) {
	switch len(args) {
	case 0:
		common.Usage("specify a branch you would like to checkout")
	case 1:
		if !refs.Exists(args[0]) {
			common.Usage(fmt.Sprintf("ref %s does not exist", args[0]))
		}
		if err := refs.PointHeadAtBranch(args[0]); err != nil {
			common.Usage(err.Error())
		}
		// Read refHash and treeHash for stats
		refHash, err := refs.ReadBranchHash(args[0])
		if err != nil {
			common.Usage(err.Error())
		}
		treeHash, err := refs.GetHeadTreeHash()
		if err != nil {
			common.Usage(err.Error())
		}
		fmt.Printf("pointed HEAD at %s\n", treeHash)
		// TODO: Implement actual checkout logic - update working dir.
		fmt.Printf("on branch master %s (commit %s)\n", args[0], refHash)
	default:
		common.Usage("Too many arguments")
	}
}
