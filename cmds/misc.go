package cmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/antoniszczepanik/gggit/objects"
	"github.com/antoniszczepanik/gggit/common"
)

func Add(args []string) {
	fmt.Println("add")
}

func Cat(args []string) {
	if len(args) != 1 {
		common.Usage("You should provide hash of object to cat.")
	}
	err := objects.PrintObject(args[0])
	if err != nil {
		fmt.Println(err)
	}
}

func Ls(args []string) {
	fmt.Println("ls-tree")
}

func Log(args []string) {
	fmt.Println("log")
}

func LsObjects(args []string) {
	objectsDir, err := common.GetGitSubdir("objects")
	if err != nil {
		common.Usage("could not find git objects dir")
	}
	dirEntries, err := os.ReadDir(objectsDir)
	if err != nil {
		common.Usage("could not read git objects dir")
	}
	for _, e := range dirEntries {
		subDirEntries, err := os.ReadDir(filepath.Join(objectsDir, e.Name()))
		if err != nil {
			common.Usage("could not read one of object subdirs")
		}
		for _, se := range subDirEntries {
			fmt.Println(e.Name() + se.Name())
		}
	}
}
