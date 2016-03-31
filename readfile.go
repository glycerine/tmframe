package tm

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// read in new-line delimited text from a file
func ReadNewlineDelimFile(path string) ([]string, error) {
	if !FileExists(path) {
		return nil, fmt.Errorf("file '%s' does not exist", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	a := make([]string, 0)

	bufIn := bufio.NewReader(f)
	for {
		lastline, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		n := len(lastline)
		if err == io.EOF && n == 0 {
			break
		}
		if n > 0 {
			if lastline[n-1] == '\n' {
				if n > 1 {
					// skip empty lines
					a = append(a, string(lastline[:n-1]))
				}
			} else {
				a = append(a, string(lastline))
			}
		}

		if err == io.EOF {
			break
		}
	}

	return a, nil
}
