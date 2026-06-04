package main

import (
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase/tools/dev/cmsqasnapshot"
)

func main() {
	if err := cmsqasnapshot.RunCreate(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
