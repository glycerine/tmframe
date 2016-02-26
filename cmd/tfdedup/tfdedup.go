package main

import (
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "%s reads through a file and deduplicates a 10 "+
		"minutes sliding window. Usage: %s <file_to_dedup>; if no file given stdin is read instead.\n",
		os.Args[0], os.Args[0])
}

func main() {
	n := len(os.Args)
	if n > 2 {
		showUse()
		os.Exit(1)
	}
	var r io.Reader
	if n == 2 {
		inputFile := os.Args[1]

		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		f, err := os.Open(inputFile)
		panicOn(err)
		r = f
	} else {
		r = os.Stdin
	}

	err := tf.Dedup(r, os.Stdout)
	if err != nil {
		panic(err)
	}
}
