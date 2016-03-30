package main

import (
	"fmt"
	tf "github.com/glycerine/tmframe"
	"io"
	"os"
	"regexp"
	"strings"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "tffilter filters raw TMFRAME streams by one or more regexes. It outputs a reduced TMFRAME stream of frames that matched all regexes. Usage: tffilter regex1 {regex2}...\n")
}

func usage(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	leftover := os.Args[1:]
	p("leftover = %v", leftover)
	if len(leftover) == 0 || (len(leftover) == 1 && strings.HasPrefix(leftover[0], "-h")) {
		fmt.Fprintf(os.Stderr, "no regex given: specify a regex to filter with.\n")
		showUse()
		os.Exit(1)
	}
	arrRegex := make([]*regexp.Regexp, 0)
	for i := range leftover {
		field := leftover[i]
		fmt.Fprintf(os.Stderr, "compiling regex %d: '%s'\n", i, field)
		re := regexp.MustCompile(field)
		arrRegex = append(arrRegex, re)
	}

	i := int64(1)

	fr := tf.NewFrameReader(os.Stdin, 1024*1024)

	var frame tf.Frame
	var raw []byte
	var err error

toploop:
	for ; err == nil; i++ {
		_, _, err, raw = fr.NextFrame(&frame)
		if err != nil {
			if err == io.EOF {
				break toploop
			}
			fmt.Fprintf(os.Stderr, "tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err)
			os.Exit(1)
		}
		str := frame.String()
		// match regex
		matchall := true
	regexLoop:
		for _, r := range arrRegex {
			if r.FindString(str) == "" {
				matchall = false
				break regexLoop
			}
		}
		if matchall {
			_, err = os.Stdout.Write(raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tffilter stopping at: '%s'", err)
			}
		}
	}
	//fmt.Fprintf(os.Stderr, "field='%s': found %v matches.\n", field, matchCount)
}
