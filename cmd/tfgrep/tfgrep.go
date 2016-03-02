package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "%s greps for fields in TMFRAME payloads. Usage: %s fieldame\n", os.Args[0], os.Args[0])
}

func usage(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

var GlobalPrettyPrint bool

func main() {
	leftover := os.Args[1:]
	//p("leftover = %v", leftover)
	if len(leftover) == 0 {
		fmt.Fprintf(os.Stderr, "no fieldname to grep for given\n")
		showUse()
		os.Exit(1)
	}
	field := leftover[0]
	fmt.Fprintf(os.Stderr, "scanning-for-escaped-json-field='%s'\n", field)
	re := regexp.MustCompile(fmt.Sprintf(`\\"%s\\":\\"([^\\]+)\\"`, field))

	bufIn := bufio.NewReader(os.Stdin)
	arr := make([]byte, 0, 1024*1024)
	buf := bytes.NewBuffer(arr)
	lineNum := int64(1)
	matchCount := int64(0)
	for {
		lastLine, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			break
		}

		if err == io.EOF && len(lastLine) == 0 {
			break
		}

		match := re.FindAllSubmatch(lastLine, -1)
		if match != nil {
			if len(match[0]) > 1 {
				for j := 1; j < len(match[0]); j++ {
					matchCount++
					fmt.Printf("%s\n", string(match[0][j]))
				}
			}
		}

		buf.Reset()
		lineNum += 1

		if err == io.EOF {
			break
		}
	}

	fmt.Fprintf(os.Stderr, "field='%s': found %v matches.\n", field, matchCount)
}
