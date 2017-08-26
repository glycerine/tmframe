/*
See https://github.com/glycerine/tmframe for the specification of the TMFRAME
format which we implement here.
*/
package tm

import (
	"bytes"
	"encoding/binary"
	"fmt"

	//needs CGO: "github.com/codahale/blake2"
	// pure Go:
	"github.com/glycerine/blake2b" // vendor https://github.com/dchest/blake2b

	"github.com/tinylib/msgp/msgp"
	"math"
	"time"
)

// PTI is the Payload Type Indicator. It is the low 3-bits
// of the Primary word in a TMFRAME message.
type PTI byte

const (
	PtiZero       PTI = 0
	PtiOneInt64   PTI = 1
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
	EvOneInt64   Evtnum = 1
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
	EvZebraPack Evtnum = 16
)

// Frame holds a fully parsed TMFRAME message.
type Frame struct {
	Prim int64 // the primary word
	//Tm()  int64 // returns low 3 bits all zeros, nanoseconds since unix epoch.
	//GetPTI() PTI   // returns low 3 bits of the primary word

	V0 float64 // primary float64 value, for EvOneFloat64 and EvTwo64

	// Ude alternatively represents V1 for EvTwo64 and EvOneInt64
	// GetV1() to access as V1.
	Ude int64 // the User-Defined-Encoding word

	// break down the Ude:
	//GetEvtnum() Evtnum
	//GetUlen()  int64

	Data []byte // the variable length payload after the UDE
}

// Tm extracts and returns the Prim timestamp from the frame (this is a UnixNano nanosecond timestamp, with the low 3 bits zeroed).
func (f *Frame) Tm() int64 {
	return f.Prim &^ 7
}

// TmTime extracts and returns the Prim timestamp from
// the frame (this is a UnixNano nanosecond timestamp,
// with the low 3 bits zeroed), then converts it to
// a UTC timezone time.Time
func (f *Frame) TmTime() time.Time {
	return time.Unix(0, f.Prim&^7).UTC()
}

// SetTm set the Prim timestamp from t. It zeros the first 3 bits of t before
// storing it, and preserves the PTI already in the primary word.
func (f *Frame) SetTm(t int64) {
	f.Prim = (t &^ 7) | f.Prim&7
}

// convert from a time.Time to a frame.Tm() comparable timestamp
func TimeToPrimTm(t time.Time) int64 {
	return t.UnixNano() &^ 7
}

// convert from a UnixNano timestamp (int64 number of nanoseconds) to a frame.Tm() comparable timestamp
func IntToPrimTm(t int64) int64 {
	return t &^ 7
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
	case PtiOneInt64:
		return 0
	case PtiOneFloat64:
		return f.V0
	case PtiTwo64:
		return f.V0
	}
	return MyNaN
}

// GetV1 retrieves the V1 value if the frame
// is of type PtiTwo64. Otherwise it returns 0.
func (f *Frame) GetV1() int64 {
	if f.GetPTI() == PtiTwo64 {
		return f.Ude
	}
	return 0
}

// SetV1 sets the Frames V1 value if the frame is
// of type PtiTwo64. Otherwise it is a no-op.
func (f *Frame) SetV1(v1 int64) {
	if f.GetPTI() == PtiTwo64 {
		f.Ude = v1
	}
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
	case PtiOneInt64:
		n = 16
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
	case PtiOneInt64:
		binary.LittleEndian.PutUint64(m[8:16], uint64(f.Ude))
	case PtiOneFloat64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
	case PtiTwo64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
		binary.LittleEndian.PutUint64(m[16:24], uint64(f.Ude))
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
// in the by []byte data. If copyData is true, we'll make a copy of
// the underlying data into the frame f.Data; otherwise we merely point
// to it. NB If the underlying buffer by is recycled/changes, and you
// want to keep around multiple frames, you should use copyData = true.
func (f *Frame) Unmarshal(by []byte, copyData bool) (rest []byte, err error) {
	// zero it all
	f.V0 = 0
	f.Ude = 0
	f.Data = []byte{}

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
	case PtiOneInt64:
		f.Ude = int64(binary.LittleEndian.Uint64(by[8:16]))
		return by[16:], nil
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
		f.Ude = int64(binary.LittleEndian.Uint64(by[16:24]))
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
			if copyData {
				cp := make([]byte, len(f.Data))
				copy(cp, f.Data)
				f.Data = cp
			}
		}
		return by[16+ucount:], nil
	default:
		panic(fmt.Sprintf("unrecog pti: %v", pti))
	}
	// panic("should never get here")
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

	if int64(len(data)) > (1<<43)-1 {
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
	q("en = %v", en)
	q("pre shift en = %b", en)
	en = en << 43
	q("post shift en = %b", en)
	q("len(data) = %v", len(data))
	q("len(data) = %b", len(data))
	var ude uint64
	if len(data) > 0 {
		// the +1 is so we zero-terminate strings -- for C bindings
		ude = uint64(len(data)+1) | en
	} else {
		ude = en
	}
	q("ude = %b", ude)

	var useData []byte
	var myUDE uint64
	//var myUlen int64

	var pti PTI
	switch evtnum {
	case EvZero:
		pti = PtiZero
	case EvOneInt64:
		pti = PtiOneInt64
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
	case EvOneInt64:
		f.Ude = v1
	case EvOneFloat64:
		f.V0 = v0
	case EvTwo64:
		f.V0 = v0
		f.Ude = v1
	}

	q("f = %#v", f)
	return f, nil
}

// String pretty prints the names of the events into a string.
func (e Evtnum) String() string {
	switch e {
	case EvErr:
		return "EvErr"
	case EvZero:
		return "EvZero"
	case EvOneInt64:
		return "EvOneInt64"
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
	return fmt.Sprintf("Ev.%d", e)
}

// String converts the Frame's header information to a string. It doesn't
// read or stingify any variable length UDE payload, even if present.
func (f Frame) String() string {
	tmu := f.Tm()
	tm := time.Unix(0, tmu).UTC()
	evtnum := f.GetEvtnum()
	ulen := f.GetUlen()

	s := fmt.Sprintf("TMFRAME %v EVTNUM %v [%v bytes] (UCOUNT %d)", tm.Format(time.RFC3339Nano), evtnum, f.NumBytes(), ulen)

	pti := f.GetPTI()

	switch pti {
	case PtiOneInt64:
		s += fmt.Sprintf(" V1:%v", f.Ude)
	case PtiOneFloat64:
		s += fmt.Sprintf(" V0:%v", f.V0)
	case PtiTwo64:
		s += fmt.Sprintf(" V0:%v V1:%v", f.V0, f.Ude)
	}

	// don't print the data; that is usually application specific.
	return s
}

// FramesEqual calls Marshal() both frames a and b and returns returns
// true if the serialized versions of a and b are byte-for-byte identical.
// FramesEqual will panics if there is a marshaling error.
func FramesEqual(a, b *Frame) bool {
	abuf, err := a.Marshal(nil)
	panicOn(err)
	bbuf, err := b.Marshal(nil)
	panicOn(err)
	return 0 == bytes.Compare(abuf, bbuf)
}

// Blake2b returns the 64-byte BLAKE2b cryptographic
// hash of the Frame. This is useful for hashing and
// de-duplicating a stream of Frames.
//
// reference: https://godoc.org/github.com/codahale/blake2
// reference: https://blake2.net/
// reference: https://tools.ietf.org/html/rfc7693
//
func (f *Frame) Blake2b() []byte {
	h, err := blake2b.New(nil)
	panicOn(err)

	n := f.NumBytes()

	var m [24]byte
	binary.LittleEndian.PutUint64(m[:8], uint64(f.Prim))
	switch {
	case n == 8:
		h.Write(m[:8])
	default:
		pti := f.GetPTI()
		switch pti {
		case PtiOneInt64:
			binary.LittleEndian.PutUint64(m[8:16], uint64(f.Ude))
			h.Write(m[:16])
		case PtiOneFloat64:
			binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
			h.Write(m[:16])
		case PtiTwo64:
			binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
			binary.LittleEndian.PutUint64(m[16:24], uint64(f.Ude))
			h.Write(m[:24])
		case PtiUDE:
			binary.LittleEndian.PutUint64(m[8:16], uint64(f.Ude))
			h.Write(m[:16])
			h.Write(f.Data)
		}
	}

	return []byte(h.Sum(nil))
}

// NewMarshalledFrame creates a frame already marshalled into the
// writeHere buffer, assuming that writeHere is large enough.
// It returns the marshalled frame as bytes, and any error.
// Effectively this is a convenience combination of NewFrame()
// followed by Marshal().
func NewMarshalledFrame(writeHere []byte, tm time.Time, evtnum Evtnum, v0 float64, v1 int64, data []byte) ([]byte, error) {

	frm, err := NewFrame(tm, evtnum, v0, v1, data)
	if err != nil {
		return nil, err
	}
	return frm.Marshal(writeHere)
}

// NewMsgpackFrame is a convenience method, taking a
// method that has had github.com/tinylib/msgp code
// generated for it. Such code will have an  msgp.Marshaler
// implementation defined by the generated code.
// The provided buf will be used if it has sufficient space,
// but is optional and can be nil.
// The marshalled frame's bytes are returned, along with
// any error encountered.
func NewMsgpackFrame(tm time.Time, m msgp.Marshaler, buf []byte) ([]byte, error) {
	bts, err := m.MarshalMsg(buf)
	if err != nil {
		return nil, err
	}
	return NewMarshalledFrame(nil, time.Now(), EvMsgpack, 0, 0, bts)
}
