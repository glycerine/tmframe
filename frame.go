package frame

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

type Frame struct {
	Prim int64 // the primary word

	V0 float64 // primary float64 value
	V1 float64 // second float64

	// breakdown the Primary
	Tm  int64 // low 3 bits all zeros, nanoseconds since unix epoch.
	Pti PTI   // low 3 bits of the primary word

	Ude int64 // the User-Defined-Encoding word

	// break down the Ude:
	IsUser bool // Q-BIT
	Evnum  Evtnum
	Ulen   int64

	Data []byte // the variable length payload after the UDE
}

func (f *Frame) Marshal(by []byte) ([]byte, error) {
	n := 8
	switch f.Pti {
	case PtiZero:
		n = 8
	case PtiOneFloat64:
		n = 16
	case PtiTwoFloat64:
		n = 24
	case PtiNaN:
		n = 8
	case PtiUDE:
		n = 16 + len(f.Data)
	case PtiNull:
		n = 8
	case PtiNA:
		n = 8
	default:
		panic(fmt.Sprintf("unrecog pti: %v", f.Pti))
	}
	m := make([]byte, n)
	binary.LittleEndian.PutUint64(m[:8], uint64(f.Prim))
	if n == 8 {
		return m, nil
	}
	switch f.Pti {
	case PtiOneFloat64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
	case PtiTwoFloat64:
		binary.LittleEndian.PutUint64(m[8:16], math.Float64bits(f.V0))
		binary.LittleEndian.PutUint64(m[16:24], math.Float64bits(f.V1))
	case PtiUDE:
		binary.LittleEndian.PutUint64(m[8:16], uint64(f.Ude))
		if n == 16 {
			return m, nil
		}
		copy(m[16:], f.Data)
	}

	return m, nil
}

var TooShortErr = fmt.Errorf("data supplied is too short to represent a TMFRAME frame")

func (f *Frame) Unmarshal(by []byte) (rest []byte, err error) {
	// zero it all
	*f = Frame{}

	n := int64(len(by))
	if n < 8 {
		return by, TooShortErr
	}
	prim := binary.LittleEndian.Uint64(by[:8])
	pti := PTI(prim % 8)

	f.Pti = pti
	f.Prim = int64(prim)
	f.Tm = int64(prim - uint64(pti))

	switch pti {
	case PtiZero:
		return by[8:], nil
	case PtiOneFloat64:
		if n < 16 {
			return by, TooShortErr
		}
		f.V0 = math.Float64frombits(binary.LittleEndian.Uint64(by[8:16]))
		return by[16:], nil
	case PtiTwoFloat64:
		if n < 24 {
			return by, TooShortErr
		}
		f.V0 = math.Float64frombits(binary.LittleEndian.Uint64(by[8:16]))
		f.V1 = math.Float64frombits(binary.LittleEndian.Uint64(by[16:24]))
		return by[24:], nil
	case PtiNaN:
		return by[8:], nil
	case PtiUDE:
		ude := binary.LittleEndian.Uint64(by[8:16])
		f.Ude = int64(ude)
		f.Evnum = Evtnum(ude >> 43)
		if f.Ude < 0 {
			Q("setting f.IsUser")
			f.IsUser = true
			f.Evnum -= (1 << 21)
		}
		ucount := ude & KeepLow43Bits
		f.Ulen = int64(ucount)
		if n < 16+f.Ulen {
			return by, TooShortErr
		}
		f.Data = by[16 : 16+ucount]
		return by[16+ucount:], nil
	case PtiNull:
		return by[8:], nil
	case PtiNA:
		return by[8:], nil
	default:
		panic(fmt.Sprintf("unrecog pti: %v", f.Pti))

	}
	panic("should never get here")
}

const KeepLow43Bits uint64 = 0x000007FFFFFFFFFF

type PTI byte

const (
	PtiZero       PTI = 0
	PtiOne        PTI = 1
	PtiOneFloat64 PTI = 2
	PtiTwoFloat64 PTI = 3
	PtiNull       PTI = 4
	PtiNA         PTI = 5
	PtiNaN        PTI = 6
	PtiUDE        PTI = 7
)

type Evtnum int32

const (
	Error   Evtnum = -1
	Zero    Evtnum = 0
	SysErr  Evtnum = 1
	Header  Evtnum = 2
	Msgpack Evtnum = 3
	Binc    Evtnum = 4
	Capnp   Evtnum = 5
	Zygo    Evtnum = 6
	Utf8    Evtnum = 7
)

func NewFrame(tm time.Time, pti PTI, evtnum Evtnum, v0 float64, v1 float64, data []byte) *Frame {
	utm := tm.UnixNano()
	mod := utm - (utm % 8)

	en := uint64(evtnum % (1 << 20))
	Q("en = %v", en)
	isUser := evtnum < 0
	Q("isUser = %v", isUser)
	if isUser {
		en |= (1 << 21)
	}
	Q("pre shift en = %b", en)
	en = en << 43
	Q("post shift en = %b", en)
	Q("len(data) = %v", len(data))
	Q("len(data) = %b", len(data))
	ude := uint64(len(data)) | en
	Q("ude = %b", ude)

	f := &Frame{
		Prim:   mod | int64(pti),
		V0:     v0,
		V1:     v1,
		Tm:     mod,
		Pti:    pti,
		Ude:    int64(ude),
		Evnum:  evtnum,
		IsUser: isUser,
		Ulen:   int64(len(data)),
		Data:   data,
	}
	Q("f = %#v", f)
	return f
}
