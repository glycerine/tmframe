package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "%s greps for escaped-json fields in TMFRAME payloads. Usage: tfcat | %s fieldname\n", os.Args[0], os.Args[0])
}

func usage(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	leftover := os.Args[1:]
	//p("leftover = %v", leftover)
	if len(leftover) == 0 || (len(leftover) == 1 && strings.HasPrefix(leftover[0], "-h")) {
		fmt.Fprintf(os.Stderr, "no fieldname given: specify a field to search for.\n")
		showUse()
		os.Exit(1)
	}
	are := make([]*regexp.Regexp, 0)
	for i := range leftover {
		field := leftover[i]
		fmt.Fprintf(os.Stderr, "scanning-for-escaped-json-field='%s'\n", field)
		re := regexp.MustCompile(fmt.Sprintf(`\\"%s\\":\\"([^\\]+)\\"`, field))
		re2 := regexp.MustCompile(fmt.Sprintf(`"%s":([^,}]+)[,}]`, field))
		are = append(are, re)
		are = append(are, re2)
	}

	bufIn := bufio.NewReader(os.Stdin)
	arr := make([]byte, 0, 1024*1024)
	buf := bytes.NewBuffer(arr)
	lineNum := int64(1)
	for {
		lastLine, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			break
		}

		if err == io.EOF && len(lastLine) == 0 {
			break
		}

		var s string
		for _, re := range are {
			match := re.FindAllSubmatch(lastLine, -1)
			if match != nil {
				if len(match[0]) > 1 {
					for j := 1; j < len(match[0]); j++ {
						s += string(match[0][j]) + "\n" // flatten matches, each to one line
					}
				}
			}
		}
		if len(s) > 0 {
			fmt.Printf(s)
		}

		buf.Reset()
		lineNum += 1

		if err == io.EOF {
			break
		}
	}

	//fmt.Fprintf(os.Stderr, "field='%s': found %v matches.\n", field, matchCount)
}
