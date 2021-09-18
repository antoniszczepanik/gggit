package main

import (
	"fmt"
	"os"

	"github.com/antoniszczepanik/gggit/cmds"
	"github.com/antoniszczepanik/gggit/utils"
)

func main() {
	if len(os.Args) < 2 {
		utils.Usage("You need to specify a gggit command.")
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "add":
		cmds.Add(args)
	case "branch":
		cmds.Branch(args)
	case "cat-file":
		cmds.Cat(args)
	case "checkout":
		cmds.Checkout(args)
	case "commit":
		cmds.Commit(args)
	case "hash-object":
		cmds.Hash(args)
	case "init":
		cmds.Init(args)
	case "log":
		cmds.Log(args)
	case "ls-tree":
		cmds.Ls(args)
	case "ls-objects":
		cmds.LsObjects(args)
	case "status":
		cmds.Status(args)
	default:
		utils.Usage(fmt.Sprintf("Command %v is not available. Did you mean sth else?\n", cmd))
	}
}
