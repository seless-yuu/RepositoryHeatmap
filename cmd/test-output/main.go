package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("This is a test output")
	fmt.Fprintf(os.Stdout, "Test to standard output\n")
	fmt.Fprintf(os.Stderr, "Test to standard error output\n")

	// Try manual flush (normally not needed)
	os.Stdout.Sync()
	os.Stderr.Sync()
}
