package mosaicmaker

import (
	"fmt"
)

func init() {

}

func Make(targetFile, srcDirectory string) int {
	fmt.Printf("Making %s from %s", targetFile, srcDirectory)
	return 0
}
