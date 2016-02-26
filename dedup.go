package tm

import (
	"fmt"
	"io"
	"os"
)

// Dedup dedups over a window of windowSize Frames a
// stream of frames from r into w.
func Dedup(r io.Reader, w io.Writer, windowSize int) error {
	fr := NewFrameReader(r, 1024*1024)
	fw := NewFrameWriter(w, 1024*1024)

	var err error
	for i := 0; err == nil; i++ {
		var frame Frame
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("dedup error from fr.NextFrame(): '%v'", err)
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

// deduper is used during dedup.
type deduper struct {
	frame *Frame
	hash  []byte
}
