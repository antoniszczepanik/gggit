package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/utils"
	"github.com/antoniszczepanik/gggit/refs"
)

func Checkout(args []string) {
	switch len(args) {
	case 0:
		utils.Usage("You should provide ref/hash you would like to checkout")
	case 1:
		if !refs.Exists(args[0]) {
			utils.Usage(fmt.Sprintf("Ref %s does not exist", args[0]))
		}
		refPath := refs.GetRefPath(args[0])
		if err := refs.PointHeadAt(refPath); err != nil {
			utils.Usage(err.Error())
		}
		refHash, err := refs.ReadHashFromRef(refPath)
		if err != nil {
			utils.Usage(err.Error())
		}
		// TODO: Implement actual checkout logic - update working dir.
		fmt.Printf("checked out %s (commit %s)\n", args[0], refHash)
	default:
		utils.Usage("Too many arguments")
	}
}
