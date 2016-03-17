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
	fmt.Fprintf(os.Stderr, "%s counts frames per sec, min, hour", os.Args[0])
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tfgroup", flag.ExitOnError)
	cfg := &tf.TfgroupConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil || cfg.Help {
		usage(err, myflags)
	}

	leftover := myflags.Args()
	//p("leftover = %v", leftover)
	if len(leftover) != 0 {
		fmt.Fprintf(os.Stderr, "tfgroup reads stdin and writes stdout, no args allowed.\n")
		showUse(myflags)
		os.Exit(1)
	}

	i := int64(1)

	f := os.Stdin
	panicOn(err)
	//buf := make([]byte, 1024*1024)
	fr := tf.NewFrameReader(f, 1024*1024)

	var frame tf.Frame

	var nextMin time.Time
	var countLastMin int64
	incr := time.Minute
	for ; err == nil; i++ {
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "tfgroup error from fr.NextFrame() at i=%v: '%v'\n", i, err)
			os.Exit(1)
		}
		cur := time.Unix(0, frame.Tm())
		if nextMin.IsZero() {
			nextMin = cur.Add(incr)
		}
		for cur.After(nextMin) {
			fmt.Printf("%v %v\n", countLastMin, nextMin)
			nextMin = nextMin.Add(incr)
			countLastMin = 0
		}
		countLastMin++
	}
	fmt.Printf("%v %v\n", countLastMin, nextMin)
}
