package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
	"time"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s replaces TMFRAME payloads with their 64-bit checksum. It reads stdin and writes stdout\n", os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfcat", flag.ExitOnError)
	cfg := &tf.TfsumConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil {
		usage(err, myflags)
	}

	leftover := myflags.Args()
	p("leftover = %v", leftover)
	if len(leftover) != 0 {
		fmt.Fprintf(os.Stderr, "tfsum reads stdin and writes stdout, no args allowed.\n")
		showUse(myflags)
		os.Exit(1)
	}

	i := int64(1)

	f := os.Stdin
	panicOn(err)
	buf := make([]byte, 1024*1024)
	fr := tf.NewFrameReader(f, 1024*1024)

	var frame tf.Frame

	for ; err == nil; i++ {
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err)
			os.Exit(1)
		}
		hash := frame.Blake2b()
		chk := int64(binary.LittleEndian.Uint64(hash[:8]))
		newf, err := tf.NewMarshalledFrame(buf, time.Unix(0, frame.Tm()), tf.EvOneInt64, 0, chk, nil)
		panicOn(err)
		_, err = os.Stdout.Write(newf)
		panicOn(err)
	}
}
