// Command demo runs small, self-contained demonstrations of the concurrency
// patterns in this repository from the command line.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
