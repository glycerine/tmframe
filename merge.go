package tm

import (
	"io"
	"sort"
)

type frameElem struct {
	frame *Frame
	bfr   *BufferedFrameReader
	index int
}

// frameSorter is used to do the merge sort
type frameSorter []*frameElem

// Len is the sorting Len function
func (p frameSorter) Len() int { return len(p) }

// Less is the sorting Less function.
func (p frameSorter) Less(i, j int) bool {
	// sort nil to end
	if p[i] == nil {
		return false
	}
	if p[j] == nil {
		return true
	}
	return p[i].frame.Tm() < p[j].frame.Tm()
}

// Swap is the sorting Swap function.
func (p frameSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Merge merges the strms input into timestamp order, based on
// the Frame.Tm() timestamp, and writes the ordered sequence
// out to the fw.Out io.writer.
func (fw *FrameWriter) Merge(strms ...*BufferedFrameReader) error {

	n := len(strms)
	peeks := make([]*frameElem, n)
	for i := 0; i < n; i++ {
		peeks[i] = &frameElem{bfr: strms[i], index: i}
	}

	var err error

	// initialize the Frames in peeks
	newlist := []*frameElem{}
	for i := range peeks {
		peeks[i].frame, err = peeks[i].bfr.Peek()
		if err != nil {
			if err == io.EOF {
				// peeks[i].frame will be nil.
				//p("err = %v, omitting peeks[i=%v] from newlist", err, i)
			} else {
				return err
			}
		} else {
			newlist = append(newlist, peeks[i])
		}
	}
	peeks = newlist

	for len(peeks) > 0 {
		if len(peeks) == 1 {
			// just copy over the rest of this stream and we're done
			//p("down to just one (%v), copying it directly over", peeks[0].index)
			_, err = peeks[0].bfr.WriteTo(fw)
			return err
		}
		// have 2 or more source left, sort and pick the earliest
		sort.Sort(frameSorter(peeks))
		// copy frame and add it to fw
		cp := *(peeks[0].frame)
		fw.Append(&cp)
		peeks[0].bfr.Advance()
		peeks[0].frame, err = peeks[0].bfr.Peek()
		if err != nil {
			if err == io.EOF {
				//p("saw err '%v',  finished with stream %v", err, peeks[0].index)
				peeks = peeks[1:]
			} else {
				return err
			}
		}
	}
	return nil
}

// Syncable allows us to sync os.File to disk, if they
// are in use in FrameStream.Out
type Syncable interface {
	Sync() error
}

// Sync writes the stream to disk, forcing any
// pending buffered writes to be persisted on disk.
func (s *FrameWriter) Sync() error {
	if s.Out == nil {
		return nil
	}
	asSync, hasSync := s.Out.(Syncable)
	if !hasSync {
		return nil
	}
	return asSync.Sync()
}
