package tm

import (
	"fmt"
	"io"
	"os"
)

//
// SkeletonDemoCopyFames is a skeleton function for
// Frame processing. Bare bones, it simply copies frames
// without doing any other transformation. It is meant
// to serve as a starting point for other customized
// processing functions.
//
// The emphasis on safety here means that this is
// deliberately not a zero copy implementation.
// Optimization is possible, but not demonstrated
// here.
//
func SkeletonDemoCopyFames(r io.Reader, w io.Writer) error {
	fr := NewFrameReader(r, 1024*1024)
	fw := NewFrameWriter(w, 1024*1024)

	var err error
	for i := 0; err == nil; i++ {
		var frame Frame
		_, _, err = fr.NextFrame(&frame)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error from fr.NextFrame(): '%v'", err)
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
