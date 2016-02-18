package tm

import (
	"fmt"
	"io"
	"os"
)

// ReadAllFrames is a helper function, reading all the
// Frames found in inputFile and returning them.
func ReadAllFrames(inputFile string) ([]*Frame, error) {
	if !FileExists(inputFile) {
		return nil, fmt.Errorf("input file '%s' does not exist.", inputFile)
	}

	var i int64
	f, err := os.Open(inputFile)
	panicOn(err)
	fr := NewFrameReader(f, 1024*1024)

	res := []*Frame{}
	for ; err == nil; i++ {
		frame := &Frame{}
		_, _, err = fr.NextFrame(frame)
		if err != nil {
			if err == io.EOF {
				return res, nil
			}
			return res, fmt.Errorf("tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err)
		}
		res = append(res, frame)
	}
	return res, nil
}
