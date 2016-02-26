package tm

import (
	"fmt"
	"io"
	"os"
)

func Dedup(r io.Reader, w io.Writer) error {
	fr := NewFrameReader(r, 1024*1024)
	fw := NewFrameWriter(w, 1024*1024)

	var err error
	for i := 0; err == nil; i++ {
		var frame Frame
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "dedup error from fr.NextFrame(): '%v'\n", err)
				os.Exit(1)
			}
		} else {
			fw.Append(&frame)
		}
		if i%1000 == 999 {
			fw.Flush()
		}
	}
	fw.Flush()
	fw.Sync()
	return nil
}

type Deduper struct {
	Frame *Frame
	Hash  []byte
}
