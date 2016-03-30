package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
	"time"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s reads a TMFRAME file and writes an index TMFRAME file.idx giving the byte offset at minute intervals. Usage: %s <file_to_index>+\n", os.Args[0], os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfindex", flag.ExitOnError)
	cfg := &tf.TfindexConfig{}
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

	i := int64(1)
nextfile:
	for _, inputFile := range leftover {
		//P("starting on inputFile '%s'", inputFile)
		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		f, err := os.Open(inputFile)
		panicOn(err)
		fr := tf.NewFrameReader(f, 1024*1024)

		writeFile := inputFile + ".idx"
		of, err := os.Create(writeFile)
		panicOn(err)

		fw := tf.NewFrameWriter(of, 1024*1024)

		var offset int64
		var frame tf.Frame
		var nextTm time.Time
		var nbytes int64

		for ; err == nil; i++ {
			_, nbytes, err, _ = fr.NextFrame(&frame)
			if err != nil {
				if err == io.EOF {
					fw.Flush()
					fw.Sync()
					of.Close()
					continue nextfile
				}
				fmt.Fprintf(os.Stderr, "tfindex error from fr.NextFrame() at i=%v: '%v'\n", i, err)
				os.Exit(1)
			}
			unix := frame.Tm()
			tm := time.Unix(0, unix)
			trunc := tm.Truncate(time.Minute)

			if i == 0 {
				first, err := tf.NewFrame(tm, tf.EvOneInt64, 0, offset, nil)
				panicOn(err)
				fw.Append(first)
				nextTm = trunc.Add(time.Minute)
			} else if tm.After(nextTm) {
				next, err := tf.NewFrame(tm, tf.EvOneInt64, 0, offset, nil)
				panicOn(err)
				fw.Append(next)
				nextTm = trunc.Add(time.Minute)
			}
			offset += nbytes
		}
	}
}
