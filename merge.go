package tm

import (
	"io"
	"sort"
)

// FrameWriter writes Frames to Out, an underlying io.Writer.
// FrameWriter may buffer frames and does not force i/o
// immediately.
type FrameWriter struct {
	Frames []*Frame
	fr     *FrameReader
	Out    io.Writer
	buf    []byte
}

// Flush writes any buffered b.Frames to b.Out.
func (b *FrameWriter) Flush() error {
	_, err := b.WriteTo(b.Out)
	return err
}

// WriteTo writes any buffered frames pending
// in the FrameWriter to w and returns the number of bytes n
// written along with any error encountered during writing.
func (b *FrameWriter) WriteTo(w io.Writer) (n int64, err error) {
	var f *Frame
	var m int
	for len(b.Frames) > 0 {
		f = b.Frames[0]
		by, err := f.Marshal(b.buf)
		if err != nil {
			return 0, err
		}
		m, err = w.Write(by)
		n += int64(m)
		if err != nil {
			return n, err
		}
		b.Frames = b.Frames[1:]
	}
	return n, nil
}

//
// Write writes len(p) bytes from p to the underlying FrameWriter.Out.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
//
func (b *FrameWriter) Write(p []byte) (n int, err error) {

	// a) first write any buffered Frames
	nn, err := b.WriteTo(b.Out)
	if err != nil {
		return int(nn), err
	}

	// b) next write any bytes in p
	m, err := b.Out.Write(p)
	nn += int64(m)
	return int(nn), err
}

// NewFrameWriter construts a new FrameWriter for buffering
// and writing Frames to w.  It imposes a
// message size limit of maxFrameBytes in order to size
// its internal marshalling buffer.
func NewFrameWriter(w io.Writer, maxFrameBytes int64) *FrameWriter {

	s := &FrameWriter{
		Out: w,
		buf: make([]byte, maxFrameBytes),
	}
	return s
}

// Append adds f the stream to be written, assuming
// it can take ownership. Copy f first if need be
// and do not write into *f after calling Append.
func (fw *FrameWriter) Append(f *Frame) {
	fw.Frames = append(fw.Frames, f)
}

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
