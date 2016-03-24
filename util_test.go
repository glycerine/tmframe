package tm

import (
	cv "github.com/glycerine/goconvey/convey"
	"testing"
	"time"
)

func Test040MakeFloat64Frames(t *testing.T) {
	cv.Convey("MakeFloat64Frames should generate a sequence of frames of equal spacing, yet omit frames with 0 specified in the delta sequence, to allow us to generate test sequences with gaps", t, func() {

		tm, err := time.Parse(time.RFC3339, "2016-03-10T00:00:00Z")
		panicOn(err)
		spacing := time.Second

		seq := []float64{-1, -5, -10, 0, 10, 0, -2}
		frames := MakeFloat64Frames(tm, spacing, seq)
		//		for i := range frames {
		//			p("frames[%v] = %s", i, *(frames[i]))
		//		}
		cv.So(len(frames), cv.ShouldEqual, 5)
		cv.So(frames[0].TmTime(), cv.ShouldResemble, tm)
		cv.So(frames[0].GetV0(), cv.ShouldResemble, float64(-1))
		cv.So(frames[1].GetV0(), cv.ShouldResemble, float64(-5))
		cv.So(frames[2].GetV0(), cv.ShouldResemble, float64(-10))
		cv.So(frames[3].GetV0(), cv.ShouldResemble, float64(10))
		cv.So(frames[4].GetV0(), cv.ShouldResemble, float64(-2))
		cv.So(frames[4].TmTime(), cv.ShouldResemble, tm.Add(6*spacing))
	})
}
