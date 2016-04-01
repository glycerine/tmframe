package tm

import (
	"fmt"
	"github.com/nats-io/gnatsd/hashmap"
	"io"
)

type DupDetectedErr struct {
	Msg string
}

func (dd *DupDetectedErr) Error() string {
	return dd.Msg
}

func NewDupDetectedErr(msg string) *DupDetectedErr {
	return &DupDetectedErr{
		Msg: msg,
	}
}

// Dedup dedups over a window of windowSize Frames a
// stream of frames from r into w. dupsW can be nil. If
// dupsW is supplied, recognized duplicate events will
// be written to this io.Writer. If detectOnly
// is true, we will return a DupDetectedErr at the
// first duplicate, to enable scanning a filesystem.
// With detectOnly set, no dedupped output Frames
// are written.
func Dedup(r io.Reader, w io.Writer, windowSize int, dupsW io.Writer, detectOnly bool) error {
	fr := NewFrameReader(r, 1024*1024)
	fw := NewFrameWriter(w, 1024*1024)

	var dupsWriter *FrameWriter
	if dupsW != nil {
		dupsWriter = NewFrameWriter(dupsW, 1024*1024)
	}

	window := make([]*dedup, windowSize)
	present := hashmap.New()

	defer func() {
		fw.Flush()
		fw.Sync()
		if dupsWriter != nil {
			dupsWriter.Flush()
			dupsWriter.Sync()
		}
	}()

	var err error
	var ptr *dedup
	for i := 0; err == nil; i++ {
		var frame Frame
		_, _, err, _ = fr.NextFrame(&frame)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("dedup error from fr.NextFrame(): '%v'", err)
			}
		} else { // err == nil

			// got a frame, check if it is a dup
			hash := frame.Blake2b()
			p := present.Get(hash)
			if p == nil {
				// not already seen
				if !detectOnly {
					fw.Append(&frame)
				}
				// memorize the new
				ptr = &dedup{count: 1, hash: hash}
				present.Set(hash, ptr)
			} else {
				// else skip duplicates, but keep a count so
				// our window is always correct. Otherwise
				// a duplicate can be missed if masked
				// by an even earlier pre-window duplicate.
				// e.g. with window size 3 and this sequence
				//  index:     0 1 2 3 4
				//  values:  [ 1 2 1 3 1 ]
				//             ^   ^   ^    <-- highlight the duplicates
				// Without the count, the dup at index 2
				// would be forgotten about when the index 0
				// value rolls out of the 'present' hash,
				// meaning that the dup at index 4 would
				// not be recognized.

				if detectOnly {
					return NewDupDetectedErr(frame.Stringify(int64(i), false, true, false))
				}
				ptr = p.(*dedup)
				ptr.count++
				if dupsWriter != nil {
					if !detectOnly {
						dupsWriter.Append(&frame)
					}
				}
			}

			if i >= windowSize {
				// deal with rolling off the last first, so
				// we have space to write
				last := present.Get(window[i%windowSize].hash).(*dedup)
				last.count--
				if last.count == 0 {
					present.Remove(last.hash)
				}
			}
			// and write our hash into our window ring
			window[i%windowSize] = ptr
		} // end else err == nil from NextFrame()

		if i%1000 == 999 {
			fw.Flush()
			if dupsWriter != nil {
				dupsWriter.Flush()
			}
		}
	}
	return nil
}

// deduper is used during dedup.
type dedup struct {
	hash  []byte
	count int
}
