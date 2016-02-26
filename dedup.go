package tm

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func showUse() {
	fmt.Fprintf(os.Stderr, "%s reads through a file and deduplicates a 10 "+
		"minutes sliding window. Usage: %s <file_to_dedup>; if no file given stdin is read instead.\n",
		os.Args[0], os.Args[0])
}

func main() {
	n := len(os.Args)
	if n > 2 {
		showUse()
		os.Exit(1)
	}
	var r io.Reader
	if n == 2 {
		inputFile := os.Args[1:]

		if !FileExists(inputFile) {
			fmt.Fprintf(os.Stderr, "input file '%s' does not exist.\n", inputFile)
			os.Exit(1)
		}

		f, err := os.Open(inputFile)
		panicOn(err)
	} else {
		r = os.Stdin
	}

	fr := NewFrameReader(r, 1024*1024)
	fw := NewFrameWriter(os.Stdout, 1024*1024)

	var err error
	for i := 0; err == nil; i++ {
		var frame tf.Frame
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "dedup error from fr.NextFrame(): '%v'\n", err)
				os.Exit(1)
			}
		} else {
			fw.Append(frame)
		}
		if i%1000 == 999 {
			fw.Flush()
		}
	}
	fw.Flush()
	fw.Sync()
}

type Deduper struct {
	Frame *Frame
	Hash  []byte
}
