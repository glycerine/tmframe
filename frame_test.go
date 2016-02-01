package frame

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
	"time"
)

func TestParsingTMFRAME(t *testing.T) {
	cv.Convey("Given that we have a timestamp and a string message, we should be able to assemble and parse the TMFRAME message.", t, func() {

		tm := time.Now()
		msg := []byte("fake msg")
		pti := PtiUDE
		utyp := Utf8
		frame := NewFrame(tm, pti, utyp, 0, 0, msg)

		nano := tm.UnixNano()
		low3 := nano % 8

		cv.So(frame.Tm, cv.ShouldEqual, nano-low3)
		cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
		cv.So(frame.Ulen, cv.ShouldEqual, len(msg))
		cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data))
		cv.So(frame.IsUser, cv.ShouldEqual, false)
		cv.So(frame.Pti, cv.ShouldEqual, PtiUDE)
		cv.So(frame.Prim, cv.ShouldEqual, frame.Tm|int64(pti))
		by, err := frame.Marshal(nil)
		panicOn(err)
		Q("by = '%v'", string(by))

		var frame2 Frame
		frame2.Unmarshal(by)
		cv.So(&frame2, cv.ShouldResemble, frame)

	})

	cv.Convey("We should be able to marshal and unmarshal User typed messages (q-bit == 1; type numbers < 0)", t, func() {

		tm := time.Now()
		msg := []byte("fake msg")
		pti := PtiUDE
		utyp := Evtnum(-34)
		frame := NewFrame(tm, pti, utyp, 0, 0, msg)

		nano := tm.UnixNano()
		low3 := nano % 8

		cv.So(frame.Tm, cv.ShouldEqual, nano-low3)
		cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
		cv.So(frame.Ulen, cv.ShouldEqual, len(msg))
		cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data))
		cv.So(frame.IsUser, cv.ShouldEqual, true)
		cv.So(frame.Pti, cv.ShouldEqual, PtiUDE)
		cv.So(frame.Prim, cv.ShouldEqual, frame.Tm|int64(pti))
		by, err := frame.Marshal(nil)
		panicOn(err)
		Q("by = '%v'", string(by))

		var frame2 Frame
		frame2.Unmarshal(by)
		cv.So(&frame2, cv.ShouldResemble, frame)

	})

}
