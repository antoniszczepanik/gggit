package cmds

import (
	"fmt"
	"os"

	"github.com/antoniszczepanik/gggit/objects"
	"github.com/antoniszczepanik/gggit/utils"
)

func Hash(args []string) {
	switch len(args) {
	case 0:
		utils.Usage("You should provide name of an entity to hash.")
	case 1:
		hash, err := hashEntityByType(args[0], false)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(hash)
	case 2:
		if args[0] != "-w" {
			utils.Usage(fmt.Sprintf("%s is not a valid option", args[0]))
			return
		}
		hash, err := hashEntityByType(args[1], true)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(hash)
	default:
		utils.Usage("Too many arguments")
	}
}

func hashEntityByType(path string, write bool) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if fileInfo.IsDir() {
		return objects.HashTree(path, write)
	}
	return hashFile(path, write)
}

func hashFile(path string, write bool) (string, error) {
	object, err := objects.CreateBlobObject(path)
	if err != nil {
		return "", err
	}
	if write {
		if err := objects.Write(object); err != nil {
			return "", err
		}
	}
	return objects.GetHash(object)
}
