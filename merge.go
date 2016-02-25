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

type FrameElem struct {
	frame *Frame
	bfr   *BufferedFrameReader
	index int
}

type FrameSorter []*FrameElem

func (p FrameSorter) Len() int { return len(p) }
func (p FrameSorter) Less(i, j int) bool {
	// sort nil to end
	if p[i] == nil {
		return false
	}
	if p[j] == nil {
		return true
	}
	return p[i].frame.Tm() < p[j].frame.Tm()
}
func (p FrameSorter) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (fw *FrameWriter) Merge(strms ...*BufferedFrameReader) error {

	n := len(strms)
	peeks := make([]*FrameElem, n)
	for i := 0; i < n; i++ {
		peeks[i] = &FrameElem{bfr: strms[i], index: i}
	}

	var err error

	// initialize the Frames in peeks
	newlist := []*FrameElem{}
	for i := range peeks {
		peeks[i].frame, err = peeks[i].bfr.Peek()
		if err != nil {
			// peeks[i].frame will be nil.
			p("err = %v, omitting peeks[i=%v] from newlist", err, i)
		} else {
			newlist = append(newlist, peeks[i])
		}
	}
	peeks = newlist

	for len(peeks) > 0 {
		if len(peeks) == 1 {
			// just copy over the rest of this stream and we're done
			p("down to just one (%v), copying it directly over", peeks[0].index)
			_, err = peeks[0].bfr.WriteTo(fw)
			return err
		}
		// have 2 or more source left, sort and pick the earliest
		sort.Sort(FrameSorter(peeks))
		// copy frame and add it to fw
		cp := *(peeks[0].frame)
		fw.Append(&cp)
		peeks[0].bfr.Advance()
		peeks[0].frame, err = peeks[0].bfr.Peek()
		if err != nil {
			p("finished with stream %v", peeks[0].index)
			peeks = peeks[1:]
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
