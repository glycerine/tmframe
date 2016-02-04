/*
See https://github.com/glycerine/tmframe for the specification of the TMFRAME
format which we implement here.
*/
package frame

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
)

// PTI is the Payload Type Indicator. It is the low 3-bits
// of the Primary word in a TMFRAME message.
type PTI byte

const (
	PtiZero       PTI = 0
	PtiOne        PTI = 1
	PtiOneFloat64 PTI = 2
	PtiTwo64      PTI = 3
	PtiNull       PTI = 4
	PtiNA         PTI = 5
	PtiNaN        PTI = 6
	PtiUDE        PTI = 7
)

// The Evtnum is the message type when pti = PtiUDE and
// UDE descriptors are in use for describing TMFRAME
// message longer than just the one Primary word.
type Evtnum int32

const (
	EvErr Evtnum = -1

	// 0-7 deliberately match the PTI to make the
	// API easier to use. Callers to NewFrame need
	// only specify an Evtnum, and the framing code
	// sets PTI and EVTNUM correctly.
	EvZero       Evtnum = 0
	EvOne        Evtnum = 1
	EvOneFloat64 Evtnum = 2
	EvTwo64      Evtnum = 3
	EvNull       Evtnum = 4
	EvNA         Evtnum = 5
	EvNaN        Evtnum = 6
	EvUDE        Evtnum = 7

	EvHeader    Evtnum = 8
	EvMsgpack   Evtnum = 9
	EvBinc      Evtnum = 10
	EvCapnp     Evtnum = 11
	EvZygo      Evtnum = 12
	EvUtf8      Evtnum = 13
	EvJson      Evtnum = 14
	EvMsgpKafka Evtnum = 15
)

// Frame holds a fully parsed TMFRAME message.
type Frame struct {
	Prim int64 // the primary word
	//GetTm()  int64 // returns low 3 bits all zeros, nanoseconds since unix epoch.
	//GetPTI() PTI   // returns low 3 bits of the primary word

	V0 float64 // primary float64 value, for EvOneFloat64 and EvTwo64
	V1 int64   // uint64 secondary payload, for EvTwo64

	Ude int64 // the User-Defined-Encoding word

	// break down the Ude:
	//GetEvtnum() Evtnum
	//GetUlen()  int64

	Data []byte // the variable length payload after the UDE
}

func (f *Frame) GetTm() int64 {
	return f.Prim &^ 7
}

func (f *Frame) GetPTI() PTI {
	return PTI(f.Prim & 7)
}

func (f *Frame) GetUDE() int64 {
	return f.Ude
}

func (f *Frame) GetUlen() int64 {
	if f.GetPTI() != PtiUDE || len(f.Data) == 0 {
		return 0
	}
	return int64(len(f.Data)) + 1 // +1 for the zero termination that only goes on the wire
}

func (f *Frame) GetEvtnum() Evtnum {
	pti := f.GetPTI()
	evnum := Evtnum(pti)
	if pti != PtiUDE {
		return evnum
	}
	evnum = Evtnum(f.Ude >> 43)
	return evnum
}

func (f *Frame) GetV0() float64 {
	pti := f.GetPTI()
	switch pti {
	case PtiZero:
		return 0
	case PtiOne:
		return 1
	case PtiOneFloat64:
		return f.V0
	case PtiTwo64:
		return f.V0
	}
	return MyNaN
}

func (f *Frame) GetV1() int64 {
	if f.GetPTI() == PtiTwo64 {
		return f.V1
	}
	return 0
}

// MyNaN provides the IEEE-754 floating point NaN value
// without having to make a call each time to math.NaN().
var MyNaN float64

func init() {
	MyNaN = math.NaN()
}

// NumBytes returns the number of bytes that the
// serialized Frame will consume on the wire. The
// count will be at least 8 bytes, and at most
// 16 + 2^43 bytes (which is 16 bytes + 8TB).
func (f *Frame) NumBytes() int64 {
	n := int64(8)
	pti := f.GetPTI()
	switch pti {
	case PtiZero:
		n = 8
	case PtiOne:
		n = 8
	case PtiOneFloat64:
		n = 16
	case PtiTwo64:
		n = 24
	case PtiNull:
		n = 8
	case PtiNA:
		n = 8
	case PtiNaN:
		n = 8
	case PtiUDE:
		n = 16
		if len(f.Data) > 0 {
			n += int64(len(f.Data)) + 1 // +1 for the zero termination that only goes on the wire
		}
	default:
		panic(fmt.Sprintf("unrecog pti: %v", pti))
	}
	return n
}

// Marshal serialized the Frame into bytes. We'll
// reuse the space pointed to by buf if there is
// sufficient space in it. We return the bytes
// that we wrote, plus any error.
func (f *Frame) Marshal(buf []byte) ([]byte, error) {
	n := f.NumBytes()

	var m []byte
	if int64(len(buf)) >= n {
		m = buf[:n]
	} else {
		m = make([]byte, n)
	}
	binary.LittleEndian.PutUint64(m[:8], uint64(f.Prim))
	if n == 8 {
		return m, nil
	}
	pti := f.GetPTI()
	switch pti {
	case PtiOneFloat64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
	case PtiTwo64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
		binary.LittleEndian.PutUint64(m[16:24], uint64(f.V1))
	case PtiUDE:
		binary.LittleEndian.PutUint64(m[8:16], uint64(f.Ude))
		if n == 16 {
			return m, nil
		}
		copy(m[16:], f.Data)
		m[n-1] = 0
	}

	return m, nil
}

// TooShortErr is returned by Frame.Unmarshal() when the
// by bytes are supplied are insufficient for the encoded
// EVTNUM or UCOUNT.
var TooShortErr = fmt.Errorf("data supplied is too short to represent a TMFRAME frame")

// Unmarshal overwrites f with the restored value of the TMFRAME found
// in the by []byte data.
func (f *Frame) Unmarshal(by []byte) (rest []byte, err error) {
	// zero it all
	*f = Frame{}

	n := int64(len(by))
	if n < 8 {
		return by, TooShortErr
	}
	prim := binary.LittleEndian.Uint64(by[:8])
	pti := PTI(prim % 8)

	f.Prim = int64(prim)

	switch pti {
	case PtiZero:
		f.V0 = 0.0
		return by[8:], nil
	case PtiOne:
		f.V0 = 1.0
		return by[8:], nil
	case PtiOneFloat64:
		if n < 16 {
			return by, TooShortErr
		}
		f.V0 = math.Float64frombits(binary.LittleEndian.Uint64(by[8:16]))
		return by[16:], nil
	case PtiTwo64:
		if n < 24 {
			return by, TooShortErr
		}
		f.V0 = math.Float64frombits(binary.LittleEndian.Uint64(by[8:16]))
		f.V1 = int64(binary.LittleEndian.Uint64(by[16:24]))
		return by[24:], nil
	case PtiNull:
		return by[8:], nil
	case PtiNA:
		return by[8:], nil
	case PtiNaN:
		// don't actually do this, as it make reflect.DeepEquals not work (of course): f.V0 = MyNaN
		return by[8:], nil
	case PtiUDE:
		ude := binary.LittleEndian.Uint64(by[8:16])
		f.Ude = int64(ude)
		ucount := ude & KeepLow43Bits
		ulen := int64(ucount)
		if n < 16+ulen {
			return by, TooShortErr
		}
		if ulen > 0 {
			f.Data = by[16 : 16+ucount-1] // -1 because the zero terminating byte only goes on the wire
		}
		return by[16+ucount:], nil
	default:
		panic(fmt.Sprintf("unrecog pti: %v", pti))
	}
	panic("should never get here")
}

// KeepLow43Bits allows one to mask off a UDE and discover
// the UCOUNT in the lower 43 bits quickly.
// For example: ucount := ude & KeepLow43Bits
//
const KeepLow43Bits uint64 = 0x000007FFFFFFFFFF

// NoDataAllowedErr is returned from NewFrame() when the
// data argument is supplied but not conveyed in that
// evtnum specified.
var NoDataAllowedErr = fmt.Errorf("data must be empty for this evtnum")

// DataTooBigErr is returned from NewFrame() if the
// user tries to submit more than 2^43 -1 bytes of data.
var DataTooBigErr = fmt.Errorf("data cannot be over 8TB - 1 byte")

// EvtnumOutOfRangeErr is retuned from NewFrame() when
// the evtnum is out of the allowed range.
var EvtnumOutOfRangeErr = fmt.Errorf("evtnum out of range. min allowed is -1048576, max is 1048575")

// Validate our acceptable range of evtnum.
// The min allowed is -1048576, max allowed is 1048575
func ValidEvtnum(evtnum Evtnum) bool {
	if evtnum > 1048575 || evtnum < -1048576 {
		return false
	}
	return true
}

// NewFrame creates a new TMFRAME message, ready to have Marshal called on
// for serialization into bytes. It will not make an internal copy of data.
// When copied on to the wire with Marshal(), a zero byte will be added
// to the data to make interop with C bindings easier; hence the UCOUNT will
// always include in its count this terminating zero byte if len(data) > 0.
//
func NewFrame(tm time.Time, evtnum Evtnum, v0 float64, v1 int64, data []byte) (*Frame, error) {

	if !ValidEvtnum(evtnum) {
		return nil, EvtnumOutOfRangeErr
	}

	if len(data) > (1<<43)-1 {
		return nil, DataTooBigErr
	}

	// sanity check that data is empty when it should be
	if len(data) > 0 {
		if evtnum >= 0 && evtnum < 7 {
			return nil, NoDataAllowedErr
		}
	}

	utm := tm.UnixNano()
	mod := utm - (utm % 8)

	en := uint64(evtnum % (1 << 21))
	Q("en = %v", en)
	Q("pre shift en = %b", en)
	en = en << 43
	Q("post shift en = %b", en)
	Q("len(data) = %v", len(data))
	Q("len(data) = %b", len(data))
	var ude uint64
	if len(data) > 0 {
		// the +1 is so we zero-terminate strings -- for C bindings
		ude = uint64(len(data)+1) | en
	} else {
		ude = en
	}
	Q("ude = %b", ude)

	var useData []byte
	var myUDE uint64
	//var myUlen int64

	var pti PTI
	switch evtnum {
	case EvZero:
		pti = PtiZero
	case EvOne:
		pti = PtiOne
	case EvOneFloat64:
		pti = PtiOneFloat64
	case EvTwo64:
		pti = PtiTwo64
	case EvNull:
		pti = PtiNull
	case EvNA:
		pti = PtiNA
	case EvNaN:
		pti = PtiNaN
	default:
		// includes case EvUDE and EvErr
		pti = PtiUDE
		useData = data
		myUDE = ude
	}

	f := &Frame{
		Prim: mod | int64(pti),
		Ude:  int64(myUDE),
		Data: useData,
	}

	// set f.V0 and v.V1
	switch evtnum {
	case EvZero:
		f.V0 = 0.0
	case EvOne:
		f.V0 = 1.0
	case EvOneFloat64:
		f.V0 = v0
	case EvTwo64:
		f.V0 = v0
		f.V1 = v1
	}

	Q("f = %#v", f)
	return f, nil
}

// FrameReader provides assistance for reading successive
// Frames from an io.Reader.
// FrameReader uses bufio to peek ahead and determine
// the size of the next frame -- see PeekNextFrame()
// and NextFrame().
type FrameReader struct {
	r             *bufio.Reader
	maxFrameBytes int64
	by            []byte
}

// NewFrameReader makes a new FrameReader. It imposes a
// message size limit of maxFrameBytes in order to size
// its internal read buffer.
func NewFrameReader(r io.Reader, maxFrameBytes int64) *FrameReader {
	return &FrameReader{
		r:             bufio.NewReaderSize(r, 16),
		maxFrameBytes: maxFrameBytes,
		by:            make([]byte, maxFrameBytes),
	}
}

// PeekNextFrame returns the size of the next frame in bytes.
//
// The returned err will be non-nil if we could encountered insufficient
// data to determine the size of the next frame. If err is
// non-nil then nBytes will be 0.
//
// Otherwise, if err is nil then nBytes holds the number of
// bytes in the next frame in FrameReader's underlying io.Reader.
func (fr *FrameReader) PeekNextFrame() (nBytes int64, err error) {

	var nAvail int64

	// peek at primary word and UDE
	by, err := fr.r.Peek(16)
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
	case PtiOne:
		return 8, nil
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

// reads the next frame and returns the number of bytes on the wire
// used by the frame. Returns a nil *Frame and nbytes of 0 if err is not nil.
func (fr *FrameReader) NextFrame() (frame *Frame, nbytes int64, err error) {
	need, err := fr.PeekNextFrame()
	if err != nil {
		return nil, 0, err
	}
	if need > fr.maxFrameBytes {
		return nil, 0, FrameTooLargeErr
	}
	if need == 0 {
		return nil, 0, io.EOF
	}

	// read 'need' number of bytes, or get an IO error
	var got int64
	var m int
	for got != need {
		m, err = fr.r.Read(fr.by[got:need])
		got += int64(m)
		if got == need {
			err = nil
			break
		}
		if err != nil {
			return nil, 0, err
		}
	}

	var f Frame
	_, err = f.Unmarshal(fr.by[:need])
	if err != nil {
		return nil, 0, err
	}
	return &f, need, nil
}

// pretty print the names of the events into a string
func (e Evtnum) String() string {
	switch e {
	case EvErr:
		return "EvErr"
	case EvZero:
		return "EvZero"
	case EvOne:
		return "EvOne"
	case EvOneFloat64:
		return "EvOneFloat64"
	case EvTwo64:
		return "EvTwo64"
	case EvNull:
		return "EvNull"
	case EvNA:
		return "EvNA"
	case EvNaN:
		return "EvNaN"
	case EvUDE:
		return "EvUDE"
	case EvHeader:
		return "EvHeader"
	case EvMsgpack:
		return "EvMsgpack"
	case EvBinc:
		return "EvBinc"
	case EvCapnp:
		return "EvCapnp"
	case EvZygo:
		return "EvZygo"
	case EvUtf8:
		return "EvUtf8"
	case EvJson:
		return "EvJson"
	case EvMsgpKafka:
		return "EvMsgpKafka"
	}
	return "unknown-Evtnum"
}

func (f *Frame) String() string {
	tmu := f.GetTm()
	tm := time.Unix(0, tmu)
	evtnum := f.GetEvtnum()
	ulen := f.GetUlen()

	s := fmt.Sprintf("TMFRAME %v EVTNUM %v [%v bytes] (UCOUNT %d)", tm.Format(time.RFC3339Nano), evtnum, f.NumBytes(), ulen)

	pti := f.GetPTI()

	switch pti {
	case PtiOneFloat64:
		s += fmt.Sprintf(" V0:%v", f.V0)
	case PtiTwo64:
		s += fmt.Sprintf(" V0:%v V1:%v", f.V0, f.V1)
	}
	// don't print the data; that is usually application specific.
	return s
}
