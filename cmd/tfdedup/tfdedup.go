package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s reads through a file and deduplicates a 10 "+
		"minutes sliding window. Usage: %s {-dupsto file} {-window size} <file_to_dedup>; if no file given stdin is read instead.\n",
		os.Args[0], os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfdedup", flag.ExitOnError)
	cfg := &tf.TfdedupConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil {
		usage(err, myflags)
	}

	leftover := myflags.Args()
	n := len(leftover)

	if n > 1 {
		usage(fmt.Errorf("too many arguments on command line"), myflags)
	}

	var r io.Reader
	var inputFile string
	if n == 1 {
		inputFile = leftover[0]

		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		f, err := os.Open(inputFile)
		panicOn(err)
		r = f
	} else {
		inputFile = "stdin"
		r = os.Stdin
	}

	var dupf *os.File
	if cfg.WriteDupsToFile != "" {
		dupf, err = os.Create(cfg.WriteDupsToFile)
		panicOn(err)
	}

	err = tf.Dedup(r, os.Stdout, cfg.WindowSize, dupf, cfg.DetectOnly)
	if cfg.DetectOnly {
		asDup, isDup := err.(*tf.DupDetectedErr)
		if isDup {
			fmt.Printf("%s has-duplicate: %s\n", inputFile, asDup)
			os.Exit(1)
		} else if err == nil {
			fmt.Printf("\n")
		}
	}
	if err != nil {
		panic(err)
	}
	os.Stdout.Sync()
	os.Stdout.Close()
}
