package main

import (
	"fmt"
	tf "github.com/glycerine/tmframe"
	"os"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "%s merges TMFRAME files. Usage: %s <file1> <file2> ...\n",
		os.Args[0], os.Args[0])
}

var GlobalPrettyPrint bool

func main() {
	if len(os.Args) <= 1 {
		showUse()
		os.Exit(1)
	}

	inputFiles := os.Args[1:]
	n := len(inputFiles)
	if n == 0 {
		fmt.Fprintf(os.Stderr, "no input files given\n")
		showUse()
		os.Exit(1)
	}

	const MB = 1024 * 1024
	outputStream := tf.NewFrameWriter(os.Stdout, MB)

	strms := make([]*tf.BufferedFrameReader, n)
	for i := 0; i < n; i++ {
		if !FileExists(inputFiles[i]) {
			fmt.Fprintf(os.Stderr, "path '%s' not found\n", inputFiles[i])
			os.Exit(1)
		}
		f, err := os.Open(inputFiles[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open path '%s': '%s'\n",
				inputFiles[i], err)
			os.Exit(1)
		}
		strms[i] = tf.NewBufferedFrameReader(f, MB, "")
	}

	// okay, now create and merge streams
	err := outputStream.Merge(strms...)
	outputStream.Sync()
	panicOn(err)
}
