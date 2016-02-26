package tm

import (
	"io"
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
