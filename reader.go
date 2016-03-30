package tm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

//////////////////////////////////////////////////
//////////////////////////////////////////////////
//
// BufferedFrameReader
//
//////////////////////////////////////////////////
//////////////////////////////////////////////////

// BufferedFrameReader supports PeekFrame(), Advance(),
// and ReadOne() that help in merging (merge sorting) two streams.
type BufferedFrameReader struct {
	Name     string
	Reader   *FrameReader
	Next     *Frame
	TmpFrame Frame
}

// NewBufferedFrameReader makes a new BufferedFrameReader. It imposes a
// message size limit of maxFrameBytes in order to size
// its internal FrameReader's buffer.
func NewBufferedFrameReader(r io.Reader, maxFrameBytes int64, name string) *BufferedFrameReader {
	s := &BufferedFrameReader{
		Name:   name,
		Reader: NewFrameReader(r, maxFrameBytes),
	}
	return s
}

// ReadOne reads the next frame and then advances
// past it. Calling it repeatedly will read all frames in
// a stream in order. ReadOne may return
// the next Frame and a non-nil error from the Advance()
// call such as io.EOF. Callers should process the
// returned *Frame (if not nil) before considering
// the returned error.
func (s *BufferedFrameReader) ReadOne() (*Frame, error) {
	if s.Next != nil {
		r := s.Next
		s.Next = nil
		return r, nil
	}
	fr, err := s.Peek()
	if err != nil {
		return fr, err
	}
	err = s.Advance()
	return fr, err
}

// Peek gets a look at the next Frame, without advancing
// past it. Repeated calls to Peek without any intervening
// ReadAndAdvance or Advance calls will return the same Frame.
func (s *BufferedFrameReader) Peek() (*Frame, error) {
	if s.Next != nil {
		return s.Next, nil
	}

	_, _, err, _ := s.Reader.NextFrame(&s.TmpFrame)
	if err != nil {
		return nil, err
	}
	s.Next = &s.TmpFrame
	return s.Next, nil
}

// Advance skips forward a frame in the stream.
// We discard the next frame --
// the next framing being the one that would have
// been returned if Peek had been called instead.
func (s *BufferedFrameReader) Advance() error {
	if s.Next != nil {
		s.Next = nil
		return nil
	}
	_, _, err, _ := s.Reader.NextFrame(&s.TmpFrame)
	if err != nil {
		return err
	}
	s.Next = &s.TmpFrame
	return nil
}

// WriteTo implements io.WriterTo. It bypasses
// Frame handling and allows copying from the underlying
// stream directly. It should be used to skip any further
// Frame processing and copy the rest of the byte stream
// directly.
func (b *BufferedFrameReader) WriteTo(w io.Writer) (n int64, err error) {
	var nn int
	if b.Next != nil {
		by, err := b.TmpFrame.Marshal(b.Reader.By)
		nn, err = w.Write(by)
		if err != nil {
			return int64(nn), err
		}
		b.Next = nil
	}
	n += int64(nn)
	m, err := b.Reader.R.WriteTo(w)
	n += m
	return n, err
}

//////////////////////////////////////////////////
//////////////////////////////////////////////////
//
// FrameReader
//
//////////////////////////////////////////////////
//////////////////////////////////////////////////

// FrameReader provides assistance for reading successive
// Frames from an io.Reader.
// FrameReader uses bufio to peek ahead and determine
// the size of the next frame -- see PeekNextFrame()
// and NextFrame().
type FrameReader struct {
	R             *bufio.Reader
	MaxFrameBytes int64
	By            []byte
}

// NewFrameReader makes a new FrameReader. It imposes a
// message size limit of maxFrameBytes in order to size
// its internal read buffer.
func NewFrameReader(r io.Reader, maxFrameBytes int64) *FrameReader {
	return &FrameReader{
		R:             bufio.NewReaderSize(r, 16),
		MaxFrameBytes: maxFrameBytes,
		By:            make([]byte, maxFrameBytes),
	}
}

// WriteTo implements io.WriterTo. It bypasses any
// Frame handling and allows copying from the underlying
// stream directly. WriteTo writes data to w until
// there is no more data to write or an error occurs.
func (b *FrameReader) WriteTo(w io.Writer) (n int64, err error) {
	return b.R.WriteTo(w)
}

// PeekNextFrameBytes returns the size of the next frame in bytes.
//
// The returned err will be non-nil if we encountered insufficient
// data to determine the size of the next frame. If err is
// non-nil then nBytes will be 0.
//
// Otherwise, if err is nil then nBytes holds the number of
// bytes in the next frame in FrameReader's underlying io.Reader.
func (fr *FrameReader) PeekNextFrameBytes() (nBytes int64, err error) {

	var nAvail int64

	// peek at primary word and UDE
	by, err := fr.R.Peek(16)
	if err != nil {
		//P("err on Peek(16): '%s'", err)
		if len(by) < 8 {
			return 0, err
		}
	}
	nAvail = int64(len(by))
	// INVAR: nAvail >= 8

	// INVAR: if nAvail < 16, then err is not nil

	// determine how many bytes this message needs
	prim := int64(binary.LittleEndian.Uint64(by[:8]))
	pti := PTI(prim % 8)

	switch pti {
	case PtiZero:
		return 8, nil
	case PtiOneInt64:
		if nAvail < 16 {
			return 0, err
		}
		return 16, nil
	case PtiNull:
		return 8, nil
	case PtiNA:
		return 8, nil
	case PtiNaN:
		return 8, nil
	case PtiOneFloat64:
		if nAvail < 16 {
			return 0, err
		}
		return 16, nil
	case PtiTwo64:
		if nAvail < 16 {
			return 0, err
		}
		return 24, nil
	case PtiUDE:
		if nAvail < 16 {
			return 0, err
		}

		ude := binary.LittleEndian.Uint64(by[8:16])
		ucount := int64(ude & KeepLow43Bits)
		return 16 + ucount, nil

	default:
		panic(fmt.Sprintf("unrecog pti: %v", pti))
	}
}

var FrameTooLargeErr = fmt.Errorf("frame was larger than FrameReader's maximum")

// NextFrame reads the next frame into fillme if provided. If fillme is
// nil, NextFrame allocates a new Frame. NextFrame returns a pointer to the filled
// frame, along with the number of bytes on the wire used by the frame.
// If err is not nil, we returns a nil *Frame and 0 for nbytes.
//
// Warning about the returned 'raw' bytes:
//
// If err is nil, the 4th return argument, raw, holds the raw bytes of the
// frame. Copy these bytes immediately if you need them, as the
// raw bytes will be overwritten on the next call to this library.
// If err is not nil, raw will be nil.
//
func (fr *FrameReader) NextFrame(fillme *Frame) (frame *Frame, nbytes int64, err error, raw []byte) {
	need, err := fr.PeekNextFrameBytes()
	if err != nil {
		return nil, 0, err, nil
	}
	if need > fr.MaxFrameBytes {
		return nil, 0, FrameTooLargeErr, nil
	}
	if need == 0 {
		return nil, 0, io.EOF, nil
	}

	// read 'need' number of bytes, or get an IO error
	var got int64
	var m int
	for got != need {
		m, err = fr.R.Read(fr.By[got:need])
		got += int64(m)
		if got == need {
			err = nil
			break
		}
		if err != nil {
			return nil, 0, err, nil
		}
	}

	yesCopyTheData := true
	if fillme == nil {
		var f Frame
		_, err = f.Unmarshal(fr.By[:need], yesCopyTheData)
		if err != nil {
			return nil, 0, err, nil
		}
		return &f, need, nil, fr.By[:need]
	}
	_, err = fillme.Unmarshal(fr.By[:need], yesCopyTheData)
	if err != nil {
		return nil, 0, err, nil
	}
	return fillme, need, nil, fr.By[:need]
}

// NextFrameBytes is like NextFrame but avoids Unmarshalling
// and so can be more efficient. NextFrameBytes reads
// the next frame into fillme if provided, but
// does not Unmarshal it; only the raw bytes of the frame are copied
// into fillme. If fillme is nil, NextFrameBytes allocates a new byte
// slice, copies the raw bytes for the next frame in, and returns it
// as nextbytes.
func (fr *FrameReader) NextFrameBytes(fillme []byte) (nextbytes []byte, err error) {
	need, err := fr.PeekNextFrameBytes()
	if err != nil {
		return nil, err
	}
	if need > fr.MaxFrameBytes {
		return nil, FrameTooLargeErr
	}
	if need == 0 {
		return nil, io.EOF
	}

	// read 'need' number of bytes, or get an IO error
	var got int64
	var m int
	for got != need {
		m, err = fr.R.Read(fr.By[got:need])
		got += int64(m)
		if got == need {
			err = nil
			break
		}
		if err != nil {
			return nil, err
		}
	}

	if len(fillme) == 0 {
		fillme = make([]byte, need)
	}
	copy(fillme, fr.By[:need])
	return fillme, nil
}
