package cmds

import (
	"fmt"

	"github.com/antoniszczepanik/gggit/common"
	"github.com/antoniszczepanik/gggit/refs"
)

func Branch(args []string) {
	switch len(args) {
	case 0:
		common.Usage("specify a branch you would like to create")
	case 1:

               if refs.Exists(args[0]) {
                       common.Usage(fmt.Sprintf("branch named '%s' already exists", args[0]))
               }
		f, err := refs.CreateNewRef(args[0])
		currentTreeHash, err := refs.GetHeadTreeHash()
		if err != nil {
			common.Usage("could not get head tree hash")
		}
		_, err = f.WriteString(currentTreeHash)
		if err != nil {
			common.Usage("could not write current tree hash to new branch ref")
		}
		fmt.Printf("created a new branch %s pointing at %s\n", args[0], currentTreeHash)
	default:
		common.Usage("Too many arguments")
	}
}
