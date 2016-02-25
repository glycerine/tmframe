package tm

import (
	"bytes"
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"io"
	"testing"
	"time"
)

func TestParsingTMFRAME(t *testing.T) {
	cv.Convey("Given that we have a timestamp and a string message, we should be able to assemble and parse the TMFRAME message.", t, func() {

		tm := time.Now()
		msg := []byte("fake msg")
		pti := PtiUDE
		event := EvUtf8
		frame, err := NewFrame(tm, event, 0, 0, msg)
		panicOn(err)

		nano := tm.UnixNano()
		low3 := nano % 8

		cv.So(frame.Tm(), cv.ShouldEqual, nano-low3)
		cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
		cv.So(frame.GetUlen(), cv.ShouldEqual, len(msg)+1)
		cv.So(frame.GetUlen(), cv.ShouldEqual, len(frame.Data)+1)
		cv.So(frame.GetPTI(), cv.ShouldEqual, PtiUDE)
		cv.So(frame.Prim, cv.ShouldEqual, frame.Tm()|int64(pti))
		by, err := frame.Marshal(nil)
		panicOn(err)
		q("by = '%v'", string(by))

		var frame2 Frame
		frame2.Unmarshal(by, false)
		cv.So(&frame2, cv.ShouldResemble, frame)

	})

	cv.Convey("We should be able to marshal and unmarshal User and system Evtnum messages", t, func() {

		x := []int{-1048576, 1048575, 1048574, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}

		for _, ev := range x {
			q("ev = %v", ev)
			tm := time.Now()
			msg := []byte("fake msg")
			if ev >= 0 && ev < 7 {
				msg = msg[:0] // don't pass data when its not used; we flag this as an error.
			}
			pti := PtiUDE
			evnum := Evtnum(ev)
			frame, err := NewFrame(tm, evnum, 0, 0, msg)
			panicOn(err)

			nano := tm.UnixNano()
			low3 := nano % 8

			cv.So(frame.Tm(), cv.ShouldEqual, nano-low3)
			if ev <= -1 || ev >= 7 {
				cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
				cv.So(frame.GetUlen(), cv.ShouldEqual, len(msg)+1) // +1 for the zero terminating byte
				cv.So(frame.GetUlen(), cv.ShouldEqual, len(frame.Data)+1)
				cv.So(frame.GetPTI(), cv.ShouldEqual, PtiUDE)
			} else {
				q(" ev = %v", ev)
				pti = PTI(ev)
				cv.So(frame.GetPTI(), cv.ShouldEqual, ev)
				cv.So(frame.Prim, cv.ShouldEqual, frame.Tm()|int64(pti))
			}
			by, err := frame.Marshal(nil)
			panicOn(err)
			q("by = '%v'", string(by))

			var frame2 Frame
			frame2.Unmarshal(by, false)
			cv.So(FramesEqual(&frame2, frame), cv.ShouldBeTrue)
			q("ev = %v and frame2.GetEvtnum()=%v", ev, frame2.GetEvtnum())
			cv.So(frame2.GetEvtnum(), cv.ShouldEqual, ev)
			q("frame2.Tm = %v", time.Unix(0, frame2.Tm()))
			q("tm = %v", tm)
		}
	})

}

func TestWrongArgsToNewFrame(t *testing.T) {
	cv.Convey("When NewFrame() is called with data and the evtnum will be ignoring it, this should be flagged as an error", t, func() {

		for ev := 0; ev < 7; ev++ {

			tm := time.Now()
			msg := []byte("fake msg")
			_, err := NewFrame(tm, Evtnum(ev), 0, 0, msg)
			cv.So(err, cv.ShouldEqual, NoDataAllowedErr)
		}

	})
}

func TestUDEwithZeroDataOkay(t *testing.T) {
	cv.Convey("When NewFrame() evtnum is UDE using and len(data) == 0, this should be okay and only result in 16 bytes serialized", t, func() {

		for ev := -15; ev < 15; ev++ {
			tm := time.Now()
			frame, err := NewFrame(tm, Evtnum(ev), 0, 0, nil)
			cv.So(err, cv.ShouldEqual, nil)
			cv.So(frame.GetUlen(), cv.ShouldEqual, 0)
			if ev < 0 {
				cv.So(frame.Ude, cv.ShouldNotEqual, 0)
				by, err := frame.Marshal(nil)
				panicOn(err)
				q("ev = %v", ev)
				cv.So(len(by), cv.ShouldEqual, 16)
			}

		}
	})
}

func TestEvtnumOutOfRange(t *testing.T) {
	cv.Convey("When NewFrame() evtnum is out of range, an error should be returned", t, func() {

		tm := time.Now()
		max := 1048575
		min := -1048576

		var err error

		_, err = NewFrame(tm, Evtnum(max), 0, 0, nil)
		cv.So(err, cv.ShouldEqual, nil)

		_, err = NewFrame(tm, Evtnum(min), 0, 0, nil)
		cv.So(err, cv.ShouldEqual, nil)

		_, err = NewFrame(tm, Evtnum(min-1), 0, 0, nil)
		cv.So(err, cv.ShouldEqual, EvtnumOutOfRangeErr)

		_, err = NewFrame(tm, Evtnum(max+1), 0, 0, nil)
		cv.So(err, cv.ShouldEqual, EvtnumOutOfRangeErr)
	})
}

// Demonstrate that the right shift does sign extension:
//
// the golang spec calls them "arithmetic shifts", but
// we known them as "sign-extending shifts".
//
// https://golang.org/ref/spec#Arithmetic_operators
//
// "The shift operators shift the left operand by the shift count
//  specified by the right operand. They implement arithmetic shifts
//  if the left operand is a signed integer and logical shifts if
//  it is an unsigned integer."
//
// Run this function if you want to see for yourself. Output:
/*
i=-8796093022208  top 21 bits: -1
i=-17592186044416  top 21 bits: -2
i=-35184372088832  top 21 bits: -4
i=-70368744177664  top 21 bits: -8
i=-140737488355328  top 21 bits: -16
i=-281474976710656  top 21 bits: -32
i=-562949953421312  top 21 bits: -64
i=-1125899906842624  top 21 bits: -128
i=-2251799813685248  top 21 bits: -256
i=-4503599627370496  top 21 bits: -512
i=-9007199254740992  top 21 bits: -1024
i=-18014398509481984  top 21 bits: -2048
i=-36028797018963968  top 21 bits: -4096
i=-72057594037927936  top 21 bits: -8192
i=-144115188075855872  top 21 bits: -16384
i=-288230376151711744  top 21 bits: -32768
i=-576460752303423488  top 21 bits: -65536
i=-1152921504606846976  top 21 bits: -131072
i=-2305843009213693952  top 21 bits: -262144
i=-4611686018427387904  top 21 bits: -524288
i=-9223372036854775808  top 21 bits: -1048576
*/
func signed_right_shift_demo() {
	b := int64(-1 << 43)
	e := int64(-1)
	for i := int64(b); i < e; i = i << 1 {
		fmt.Printf("i=%v  top 21 bits: %v\n", i, int32(i>>43))
	}
}

func Test200FrameReader(t *testing.T) {
	cv.Convey("Given that consumers of TMFRAME frames would like to peek ahead 1-2 words to discover the length of the next message in a stream, so they can allocate message buffers, the FrameReader should wrap and buffer an io.Reader to provide PeekNextFrame.", t, func() {

		tm := time.Now()
		//max := 1048575
		var err error
		//var by []byte

		p("FrameReader.ReadNextFrame() should return io.EOF or empty []byte / stream at end of file")
		var empty bytes.Buffer
		fr := NewFrameReader(&empty, 64*1024)
		nBytes, err := fr.PeekNextFrameBytes()
		cv.So(nBytes, cv.ShouldEqual, 0)
		cv.So(err, cv.ShouldEqual, io.EOF)
		//		by, err = fr.ReadNextFrame(empty)
		//		cv.So(len(by), cv.ShouldEqual, 0)
		//		cv.So(err, cv.ShouldEqual, io.EOF)

		p("FrameReader.PeekNextFrame() should return the size without consuming the Frame")
		f8, err := NewFrame(tm, EvZero, 0, 0, nil)
		panicOn(err)
		b8, err := f8.Marshal(nil)
		panicOn(err)
		buf8 := bytes.NewBuffer(b8)
		fr = NewFrameReader(buf8, 64*1024)
		nBytes, err = fr.PeekNextFrameBytes()
		cv.So(nBytes, cv.ShouldEqual, 8)
		cv.So(err, cv.ShouldEqual, nil)

		/*
			f16, err := NewFrame(tm, EvOneFloat64, 0, 0, nil)
			panicOn(err)
			b16 := bytes.NewBuffer(f16)
			cv.So(PeekNextFrame(b16), cv.ShouldEqual, 16)

			f24, err := NewFrame(tm, EvTwo64, 0, 0, nil)
			panicOn(err)
			b24 := bytes.NewBuffer(f24)
			cv.So(PeekNextFrame(b24), cv.ShouldEqual, 24)

			f36, err := NewFrame(tm, EvMsgpKafka, 0, 0, make([]byte, 20))
			panicOn(err)
			b36 := bytes.NewBuffer(f36)
			cv.So(PeekNextFrame(b36), cv.ShouldEqual, 36)
		*/
	})
}
