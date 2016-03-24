package tm

import (
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"io"
	"os"
	"testing"
	"time"
)

func Test010InForceAtReturnsFrameBefore(t *testing.T) {

	cv.Convey(`Given a Series s, the call s.LastInForceBefore(tm) should `+
		`return the last Frame strictly before tm`, t, func() {

		outpath := "test.frames.out.1"
		testFrames, tms, by := GenTestFrames(5, &outpath)

		// chuck unmarshal of the generated frames while we're at it
		rest := by
		var err error
		for i := range testFrames {
			var newFr Frame
			rest, err = newFr.Unmarshal(rest, false)
			panicOn(err)
			if !FramesEqual(&newFr, testFrames[i]) {
				panic(fmt.Sprintf("frame %v error: expected '%s' to equal '%s' upon unmarshal, but did not.", i, newFr, testFrames[i]))
			}
		}

		// read back from file, to emulate actual use.
		f, err := os.Open(outpath)
		panicOn(err)
		fr := NewFrameReader(f, 1024*1024)

		var frame Frame
		i := 0
		for ; err == nil; i++ {
			_, _, err = fr.NextFrame(&frame)
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(fmt.Sprintf("tfcat error from fr.NextFrame() at i=%v: '%v'\n", i, err))
			}
		}

		sers := NewSeriesFromFrames(testFrames)
		at, status, i := sers.LastInForceBefore(tms[2])
		//P("at, status = %v, %v", at, status)
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(time.Unix(0, at.Tm()).UTC(), cv.ShouldResemble, tms[1])
		cv.So(i, cv.ShouldEqual, 1)

		at, status, i = sers.LastInForceBefore(tms[4].Add(time.Hour))
		//P("at, status = %v, %v", at, status)
		cv.So(status, cv.ShouldEqual, InFuture)
		cv.So(time.Unix(0, at.Tm()).UTC(), cv.ShouldResemble, tms[4])
		cv.So(i, cv.ShouldEqual, 4)

		at, status, i = sers.LastInForceBefore(tms[0])
		//P("at, status = %v, %v", at, status)
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(at, cv.ShouldEqual, nil)
		cv.So(i, cv.ShouldEqual, -1)
	})

	cv.Convey(`Given a Series ser, the call ser.FirstInForceBefore(t) should `+
		`return the first Frame in any series of repeated frames tied at the`+
		` same timestamp s, when s < t and no other timestamp r on an existing`+
		` frame lives between them; there is no r: s < r < t`, t, func() {

		repeat, tms, _ := GenTestFramesSequence(5, nil)

		// have the middle 3 repeat the same time;
		for i := 1; i < 4; i++ {
			repeat[i].SetTm(repeat[3].Tm())
		}

		sers := NewSeriesFromFrames(repeat)

		at, status, i := sers.FirstInForceBefore(tms[4])
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(at.GetV0(), cv.ShouldEqual, 1)
		cv.So(i, cv.ShouldEqual, 1)

		p("FristInForceBefore InFuture test")
		at, status, i = sers.FirstInForceBefore(tms[4].Add(time.Hour))
		cv.So(status, cv.ShouldEqual, InFuture)
		cv.So(at.GetV0(), cv.ShouldEqual, 4)
		cv.So(i, cv.ShouldEqual, 4)

		sers.Frames = sers.Frames[:4]
		at, status, i = sers.FirstInForceBefore(tms[4].Add(time.Hour))
		cv.So(status, cv.ShouldEqual, InFuture)
		cv.So(at.GetV0(), cv.ShouldEqual, 1)
		cv.So(i, cv.ShouldEqual, 1)

		at, status, i = sers.FirstInForceBefore(tms[0].Add(-time.Nanosecond))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(at, cv.ShouldEqual, nil)
		cv.So(i, cv.ShouldEqual, -1)
	})

	cv.Convey(`FirstAtOrBefore(tm) returns `+
		`the first of (any repeated time timestamps) precisely time matched tick if available at exactly tm, `+
		` and otherwise returns the Frame (if any) before tm`,
		t, func() {

			outpath := "test.frames.out.2"
			repeat, tms, _ := GenTestFramesSequence(5, &outpath)

			// have them repeat the same time but with different values 0..4
			// so we can distinguish if the first one was returned.
			for i := 0; i < 5; i++ {
				repeat[i].SetTm(repeat[0].Tm())
			}

			sers := NewSeriesFromFrames(repeat)

			at, status, i := sers.FirstAtOrBefore(tms[0])
			cv.So(status, cv.ShouldEqual, Avail)
			cv.So(at.GetV0(), cv.ShouldEqual, 0)
			cv.So(i, cv.ShouldEqual, 0)

			p("FristAtOrBefore InFuture test")
			at, status, i = sers.FirstAtOrBefore(tms[0].Add(time.Hour))
			cv.So(status, cv.ShouldEqual, InFuture)
			cv.So(at.GetV0(), cv.ShouldEqual, 0)
			cv.So(i, cv.ShouldEqual, 0)

			at, status, i = sers.FirstAtOrBefore(tms[0].Add(-time.Nanosecond))
			cv.So(status, cv.ShouldEqual, InPast)
			cv.So(at, cv.ShouldEqual, nil)
			cv.So(i, cv.ShouldEqual, -1)

		})

	cv.Convey(`LastAtOrBefore(tm) returns `+
		`the last of (any repeat time timestamped) precisely time matched tick if available at exactly tm, `+
		` and otherwise returns the Frame (if any) before tm`,
		t, func() {

			outpath := "test.frames.out.3"
			repeat, _, _ := GenTestFramesSequence(5, &outpath)

			// have them repeat the same time but with different values 0..4
			// so we can distinguish if the first one was returned.
			for i := 0; i < 5; i++ {
				repeat[i].SetTm(repeat[0].Tm())
			}

			sers := NewSeriesFromFrames(repeat)

			base := time.Unix(0, repeat[0].Tm())
			at, status, i := sers.LastAtOrBefore(base)
			cv.So(status, cv.ShouldEqual, Avail)
			cv.So(at.GetV0(), cv.ShouldEqual, 4)
			cv.So(i, cv.ShouldEqual, 4)

			p("LastAtOrBefore InFuture test")
			at, status, i = sers.LastAtOrBefore(base.Add(time.Hour))
			cv.So(status, cv.ShouldEqual, InFuture)
			cv.So(at.GetV0(), cv.ShouldEqual, 4)
			cv.So(i, cv.ShouldEqual, 4)

			at, status, i = sers.LastAtOrBefore(base.Add(-time.Nanosecond))
			cv.So(status, cv.ShouldEqual, InPast)
			cv.So(at, cv.ShouldEqual, nil)
			cv.So(i, cv.ShouldEqual, -1)

		})

}

func Test015ExtendedRepetitionTestLastInForceBefore(t *testing.T) {

	cv.Convey(`Given a Series s, the call s.LastInForceBefore(tm) should `+
		`return the last repeated Frame < tm, even with varying repetition patterns`, t, func() {
		reps := []int{5, 5, 5, 5}
		sers := GenerateSeriesWithRepeats(reps)

		_, status, i := sers.LastInForceBefore(time.Unix(0, sers.Frames[19].Tm()+10))
		cv.So(i, cv.ShouldEqual, 19)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[19].Tm()))
		cv.So(i, cv.ShouldEqual, 14)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[14].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 9)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[9].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 4)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 2, 1, 2}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[5].Tm()+10))
		cv.So(i, cv.ShouldEqual, 5)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[5].Tm()))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 3)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 2)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1, 1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[3].Tm()+10))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[0].Tm()+10))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[1].Tm()+10))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)
	})
}

func Test016ExtendedRepetitionTestLastAtOrBefore(t *testing.T) {

	cv.Convey(`Given a Series s, the call s.LastAtOrBefore(tm) should `+
		`return the last repeated Frame <= tm, even with varying repetition patterns`, t, func() {
		reps := []int{5, 5, 5, 5}
		sers := GenerateSeriesWithRepeats(reps)

		_, status, i := sers.LastAtOrBefore(time.Unix(0, sers.Frames[19].Tm()+10))
		cv.So(i, cv.ShouldEqual, 19)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[15].Tm()))
		cv.So(i, cv.ShouldEqual, 19)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[10].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 14)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[5].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 9)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 4)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 2, 1, 2}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[5].Tm()+10))
		cv.So(i, cv.ShouldEqual, 5)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[5].Tm()))
		cv.So(i, cv.ShouldEqual, 5)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 5)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 3)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 2)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 2)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1, 1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[3].Tm()+10))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(i, cv.ShouldEqual, 2)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()+10))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[1].Tm()+10))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.LastAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

	})
}

// 017

func Test017ExtendedRepetitionTestFirstAtOrBefore(t *testing.T) {

	cv.Convey(`Given a Series s, the call s.FirstAtOrBefore(tm) should `+
		`return the first repeat <= tm, even with varying repetition patterns`, t, func() {
		reps := []int{5, 5, 5, 5}
		sers := GenerateSeriesWithRepeats(reps)

		_, status, i := sers.FirstAtOrBefore(time.Unix(0, sers.Frames[19].Tm()+10))
		cv.So(i, cv.ShouldEqual, 15)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[15].Tm()))
		cv.So(i, cv.ShouldEqual, 15)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[14].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 10)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[9].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 5)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 2, 1, 2}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[5].Tm()+10))
		cv.So(i, cv.ShouldEqual, 4)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[5].Tm()))
		cv.So(i, cv.ShouldEqual, 4)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 4)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 3)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 1)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 1)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1, 1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[3].Tm()+10))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(i, cv.ShouldEqual, 2)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()+10))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[1].Tm()+10))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstAtOrBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

	})
}

// 018

func Test018ExtendedRepetitionTestFirstInForceBefore(t *testing.T) {

	cv.Convey(`Given a Series s, the call s.FirstInForceBefore(tm) should `+
		`return the first of all repeats at the nearest point strictly < tm, even with varying repetition patterns`, t, func() {
		reps := []int{5, 5, 5, 5}
		sers := GenerateSeriesWithRepeats(reps)

		_, status, i := sers.FirstInForceBefore(time.Unix(0, sers.Frames[19].Tm()+10))
		cv.So(i, cv.ShouldEqual, 15)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[19].Tm()))
		cv.So(i, cv.ShouldEqual, 10)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[14].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 5)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[9].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 2, 1, 2}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[5].Tm()+10))
		cv.So(i, cv.ShouldEqual, 4)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[5].Tm()))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[4].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 3)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 1)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(i, cv.ShouldEqual, 0)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1, 1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[3].Tm()+10))
		cv.So(i, cv.ShouldEqual, 3)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[3].Tm()))
		cv.So(i, cv.ShouldEqual, 2)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[2].Tm()))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, -1)
		cv.So(status, cv.ShouldEqual, InPast)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()+10))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, -1)
		cv.So(status, cv.ShouldEqual, InPast)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()-10))
		cv.So(status, cv.ShouldEqual, InPast)
		cv.So(i, cv.ShouldEqual, -1)

		reps = []int{1, 1}
		sers = GenerateSeriesWithRepeats(reps)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[1].Tm()+10))
		cv.So(i, cv.ShouldEqual, 1)
		cv.So(status, cv.ShouldEqual, InFuture)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[1].Tm()))
		cv.So(i, cv.ShouldEqual, 0)
		cv.So(status, cv.ShouldEqual, Avail)

		_, status, i = sers.FirstInForceBefore(time.Unix(0, sers.Frames[0].Tm()))
		cv.So(i, cv.ShouldEqual, -1)
		cv.So(status, cv.ShouldEqual, InPast)

	})
}
