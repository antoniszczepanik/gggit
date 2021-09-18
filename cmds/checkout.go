package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/utils"
	"github.com/antoniszczepanik/gggit/refs"
)

func Checkout(args []string) {
	switch len(args) {
	case 0:
		utils.Usage("specify a branch you would like to checkout")
	case 1:
		if !refs.Exists(args[0]) {
			utils.Usage(fmt.Sprintf("ref %s does not exist", args[0]))
		}
		refPath := refs.GetBranchRefPath(args[0])
		if err := refs.PointHeadAt(refPath); err != nil {
			utils.Usage(err.Error())
		}
		// Read refHash and treeHash for stats
		refHash, err := refs.ReadHashFromRef(refPath)
		if err != nil {
			utils.Usage(err.Error())
		}
		treeHash, err := refs.GetHeadTreeHash()
		if err != nil {
			utils.Usage(err.Error())
		}
		fmt.Printf("pointed HEAD at %s\n", treeHash)
		// TODO: Implement actual checkout logic - update working dir.
		fmt.Printf("on branch master %s (commit %s)\n", args[0], refHash)
	default:
		utils.Usage("Too many arguments")
	}
}
