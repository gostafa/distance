// Command distance reports each Go package's distance from the main sequence.
//
//	distance [flags] [patterns...]
package main

import (
	"os"

	"github.com/gostafa/distance/internal/cli"
)

var (
	run  = cli.Run
	exit = os.Exit
)

func main() {
	exit(run(os.Args[1:]))
}
