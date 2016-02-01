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
		event := EvUtf8
		frame := NewFrame(tm, event, 0, 0, msg)

		nano := tm.UnixNano()
		low3 := nano % 8

		cv.So(frame.Tm, cv.ShouldEqual, nano-low3)
		cv.So(string(frame.Data), cv.ShouldEqual, string(msg)+string(0))
		cv.So(frame.Ulen, cv.ShouldEqual, len(msg)+1)
		cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data))
		cv.So(frame.Pti, cv.ShouldEqual, PtiUDE)
		cv.So(frame.Prim, cv.ShouldEqual, frame.Tm|int64(pti))
		by, err := frame.Marshal(nil)
		panicOn(err)
		Q("by = '%v'", string(by))

		var frame2 Frame
		frame2.Unmarshal(by)
		cv.So(&frame2, cv.ShouldResemble, frame)

	})

	cv.Convey("We should be able to marshal and unmarshal User typed messages (q-bit == 1; with evnnum < 0)", t, func() {

		for ev := -4; ev < 14; ev++ {

			tm := time.Now()
			msg := []byte("fake msg")
			pti := PtiUDE
			evnum := Evtnum(ev)
			frame := NewFrame(tm, evnum, 0, 0, msg)

			nano := tm.UnixNano()
			low3 := nano % 8

			cv.So(frame.Tm, cv.ShouldEqual, nano-low3)
			if ev <= -1 || ev >= 7 {
				cv.So(string(frame.Data), cv.ShouldEqual, string(msg)+string(0))
				cv.So(frame.Ulen, cv.ShouldEqual, len(msg)+1) // +1 for the zero terminating byte
				cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data))
				cv.So(frame.Pti, cv.ShouldEqual, PtiUDE)
			} else {
				Q(" ev = %v", ev)
				pti = PTI(ev)
				cv.So(frame.Pti, cv.ShouldEqual, ev)
				cv.So(frame.Prim, cv.ShouldEqual, frame.Tm|int64(pti))
			}
			by, err := frame.Marshal(nil)
			panicOn(err)
			Q("by = '%v'", string(by))

			var frame2 Frame
			frame2.Unmarshal(by)
			cv.So(&frame2, cv.ShouldResemble, frame)
			cv.So(frame2.Evnum, cv.ShouldEqual, ev)
			Q("frame2.Tm = %v", time.Unix(0, frame2.Tm))
			Q("tm = %v", tm)
		}
	})

}
