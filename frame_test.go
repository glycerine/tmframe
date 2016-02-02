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
		frame, err := NewFrame(tm, event, 0, 0, msg)
		panicOn(err)

		nano := tm.UnixNano()
		low3 := nano % 8

		cv.So(frame.GetTm(), cv.ShouldEqual, nano-low3)
		cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
		cv.So(frame.Ulen, cv.ShouldEqual, len(msg)+1)
		cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data)+1)
		cv.So(frame.GetPTI(), cv.ShouldEqual, PtiUDE)
		cv.So(frame.Prim, cv.ShouldEqual, frame.GetTm()|int64(pti))
		by, err := frame.Marshal(nil)
		panicOn(err)
		Q("by = '%v'", string(by))

		var frame2 Frame
		frame2.Unmarshal(by)
		cv.So(&frame2, cv.ShouldResemble, frame)

	})

	cv.Convey("We should be able to marshal and unmarshal User and system Evtnum messages", t, func() {

		for ev := -4; ev < 14; ev++ {
			Q("ev = %v", ev)
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

			cv.So(frame.GetTm(), cv.ShouldEqual, nano-low3)
			if ev <= -1 || ev >= 7 {
				cv.So(string(frame.Data), cv.ShouldEqual, string(msg))
				cv.So(frame.Ulen, cv.ShouldEqual, len(msg)+1) // +1 for the zero terminating byte
				cv.So(frame.Ulen, cv.ShouldEqual, len(frame.Data)+1)
				cv.So(frame.GetPTI(), cv.ShouldEqual, PtiUDE)
			} else {
				Q(" ev = %v", ev)
				pti = PTI(ev)
				cv.So(frame.GetPTI(), cv.ShouldEqual, ev)
				cv.So(frame.Prim, cv.ShouldEqual, frame.GetTm()|int64(pti))
			}
			by, err := frame.Marshal(nil)
			panicOn(err)
			Q("by = '%v'", string(by))

			var frame2 Frame
			frame2.Unmarshal(by)
			cv.So(&frame2, cv.ShouldResemble, frame)
			cv.So(frame2.Evnum, cv.ShouldEqual, ev)
			Q("frame2.Tm = %v", time.Unix(0, frame2.GetTm()))
			Q("tm = %v", tm)
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
			cv.So(frame.Ulen, cv.ShouldEqual, 0)
			if ev < 0 {
				cv.So(frame.Ude, cv.ShouldNotEqual, 0)
				by, err := frame.Marshal(nil)
				panicOn(err)
				Q("ev = %v", ev)
				cv.So(len(by), cv.ShouldEqual, 16)
			}

		}
	})
}
