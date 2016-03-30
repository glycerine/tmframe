package main

import (
	"flag"
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
	"regexp"
	"strings"
)

func showUse(myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "tffilter filters raw TMFRAME streams on stdin by one or more regexes. It writes to stdout a reduced TMFRAME stream of frames that matched all regexes. Usage: tffilter regex1 {regex2}...\n")
	myflags.PrintDefaults()
}

func usage(err error, myflags *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	showUse(myflags)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	myflags := flag.NewFlagSet("tffilter", flag.ExitOnError)
	cfg := &tf.TffilterConfig{}
	cfg.DefineFlags(myflags)

	err := myflags.Parse(os.Args[1:])
	err = cfg.ValidateConfig()
	if err != nil || cfg.Help {
		usage(err, myflags)
	}

	leftover := myflags.Args()

	regs := leftover
	//p("regs = %v", regs)
	if cfg.RegexFile != "" {
		regs, err = tf.ReadNewlineDelimFile(cfg.RegexFile)
		if err != nil {
			usage(err, myflags)
		}
		if len(regs) == 0 {
			fmt.Fprintf(os.Stderr, "no regex given from file '%s': specify at least one regex to filter with.\n", cfg.RegexFile)
			showUse(myflags)
			os.Exit(1)
		}
	} else {
		if len(regs) == 0 || (len(regs) == 1 && strings.HasPrefix(regs[0], "-h")) {
			fmt.Fprintf(os.Stderr, "no regex given: specify at least one regex to filter with.\n")
			showUse(myflags)
			os.Exit(1)
		}
	}

	// INVAR: regs specified, and len(regs) > 0

	arrRegex := make([]*regexp.Regexp, 0)
	for i := range regs {
		field := regs[i]
		//fmt.Fprintf(os.Stderr, "compiling regex %d: '%s'\n", i, field)
		re := regexp.MustCompile(field)
		arrRegex = append(arrRegex, re)
	}

	i := int64(1)

	fr := tf.NewFrameReader(os.Stdin, 1024*1024)

	var frame tf.Frame
	var raw []byte
	n := len(regs)

toploop:
	for ; err == nil; i++ {
		_, _, err, raw = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				break toploop
			}
			fmt.Fprintf(os.Stderr, "tffilter error from fr.NextFrame() at i=%v: '%v'\n", i, err)
			os.Exit(1)
		}
		str := frame.Stringify(-1, false, false)
		// match regex
		matchN := 0
		for _, r := range arrRegex {
			o := r.FindString(str)
			//fmt.Fprintf(os.Stderr, "tffilter at i=%v, matching frame '%s' against regex '%s': output is: '%s'\n", j, str, regs[j], o)
			if o != "" {
				matchN++
			}
		}
		switch {
		case !cfg.ExcludeMatches:
			// regex all must match for us to let the frame through the filters
			if matchN == n {
				_, err = os.Stdout.Write(raw)
				if err != nil {
					fmt.Fprintf(os.Stderr, "tffilter stopping at: '%s'", err)
				}
			}
		case cfg.ExcludeMatches:
			// under -x, we only *exclude* when all the filters match
			if matchN < n {
				_, err = os.Stdout.Write(raw)
				if err != nil {
					fmt.Fprintf(os.Stderr, "tffilter stopping at: '%s'", err)
				}
			}
		}
	}
	//fmt.Fprintf(os.Stderr, "field='%s': found %v matches.\n", field, matchCount)
}
