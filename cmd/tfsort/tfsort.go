package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	"os"
	"sort"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s sorts each TMFRAME `file` into temp file `file.sorted`. Then files are merged and written to stdout. Temp files are deleted unless -k is given. Usage: %s {-k} <file_to_sort>+\n", os.Args[0], os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfsort", flag.ExitOnError)
	cfg := &tf.TfsortConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil {
		usage(err, myflags)
	}

	leftover := myflags.Args()
	//Q("leftover = %v", leftover)
	if len(leftover) == 0 {
		fmt.Fprintf(os.Stderr, "no input files given\n")
		showUse(myflags)
		os.Exit(1)
	}

	wrote := []*os.File{}
	wroteTmp := []string{}
	for _, inputFile := range leftover {
		//P("starting on inputFile '%s'", inputFile)
		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		frames, err := tf.ReadAllFrames(inputFile)
		panicOn(err)

		sort.Stable(tf.TimeSorter(frames))

		writeFile := inputFile + ".sorted"
		of, err := os.Create(writeFile)
		panicOn(err)
		wrote = append(wrote, of)
		wroteTmp = append(wroteTmp, writeFile)

		fw := tf.NewFrameWriter(of, 1024*1024)
		fw.Frames = frames
		_, err = fw.WriteTo(of)
		panicOn(err)
		fw.Sync()
		of.Seek(0, 0)
	}

	// INVAR: individual files are sorted, now merge to stdout

	const MB = 1024 * 1024
	outputStream := tf.NewFrameWriter(os.Stdout, MB)

	strms := make([]*tf.BufferedFrameReader, len(wrote))
	for i := range wrote {
		strms[i] = tf.NewBufferedFrameReader(wrote[i], MB)
	}

	// okay, now create and merge streams
	err = outputStream.Merge(strms...)
	outputStream.Sync()
	panicOn(err)

	if !cfg.KeepTmpFiles {
		for _, w := range wroteTmp {
			os.Remove(w)
		}
	}

}
