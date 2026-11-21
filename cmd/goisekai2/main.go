// Command goisekai2 is the daemon entry point. Greenfield: only the version
// surface exists until the first OpenSpec change lands.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("goisekai2", version)
		return
	}
	fmt.Fprintln(os.Stderr, "goisekai2: no commands implemented yet (see openspec/changes)")
	os.Exit(2)
}
