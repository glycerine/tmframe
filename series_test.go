package tm

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	cv "github.com/glycerine/goconvey/convey"
	"io"
	"os"
	"testing"
	"time"
)

func Test011MovingAverage(t *testing.T) {

	cv.Convey(`Given a primary series of TMFRAMEs, we should be able to generate a derived series representing the moving average`, t, func() {

		outpath := "test.seq.frames.out"
		//testFrames, tms, by := GenTestFramesSequence(10, &outpath)
		GenTestFramesSequence(10, &outpath)

		//ser := Series(testFrames)

	})
}

func Test010InForceAtReturnsFrameBefore(t *testing.T) {

	//cv.Convey(`Given an Index, searching with JustBefore(tm) should return the Frame in force at tm`, t, func() {
	cv.Convey(`Given an set of tm.Frames, a Ring holding those frame bytes, and an Index on that Ring,`+
		` when we add messages to the ring, the ability to locate the event at a`+
		` given timestamp with Index.InForceAt() should be preserved.`, t, func() {

		outpath := "test.frames.out"
		testFrames, tms, by := GenTestFrames(5, &outpath)

		// chuck unmarshal of the generated frames while we're at it
		rest := by
		var err error
		for i := range testFrames {
			var newFr Frame
			rest, err = newFr.Unmarshal(rest)
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
		at, status := sers.InForceBefore(tms[2])
		P("at, status = %v, %v", at, status)
		cv.So(status, cv.ShouldEqual, Avail)
		cv.So(time.Unix(0, at.Tm()).UTC(), cv.ShouldResemble, tms[1])
		/*testFrames, tms, by := GenTestFrames(5, &outpath)
		P("by=%#v", by)
		ring := NewFrameRing(5)
		for i := range testFrames {
			//ring.Add(testFrames[i])
		}
		*/
		P("\n tms:%#v  \n\n testFrames:%#v \n\n", tms, testFrames)
		//cv.So(x.SexpString(), cv.ShouldEqual, ` (snoopy chld: (hellcat speed:567))`)
	})
}

// generate n test Frames, with 4 different frame types, and randomly varying sizes
// if outpath is non-nill, write to that file.
func GenTestFrames(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		switch i % 3 {
		case 0:
			// generate a random length message payload
			m := cryptoRandNonNegInt() % 254
			data := make([]byte, m)
			for j := 0; j < m; j++ {
				data[j] = byte(j)
			}
			f0, err = NewFrame(t, EvMsgpKafka, 0, 0, data)
			panicOn(err)
		case 1:
			f0, err = NewFrame(t, EvZero, 0, 0, nil)
			panicOn(err)
		case 2:
			f0, err = NewFrame(t, EvTwo64, float64(i), int64(i), nil)
			panicOn(err)
		case 3:
			f0, err = NewFrame(t, EvOneFloat64, float64(i), 0, nil)
			panicOn(err)
		}
		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)
		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		f, err := os.Create(*outpath)
		panicOn(err)
		_, err = f.Write(by)
		panicOn(err)
		f.Close()
	}

	return
}

func cryptoRandNonNegInt() int {
	b := make([]byte, 8)
	_, err := cryptorand.Read(b)
	panicOn(err)
	r := int(binary.LittleEndian.Uint64(b))
	if r < 0 {
		return -r
	}
	return r
}

// generate 0..(n-1) as floating point EvOneFloat64 frames
func GenTestFramesSequence(n int, outpath *string) (frames []*Frame, tms []time.Time, by []byte) {

	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := 0; i < n; i++ {
		t := t0.Add(time.Second * time.Duration(i))
		tms = append(tms, t)
		f0, err = NewFrame(t, EvOneFloat64, float64(i), 0, nil)
		panicOn(err)
		frames = append(frames, f0)
		b0, err := f0.Marshal(nil)
		panicOn(err)
		by = append(by, b0...)
	}

	if outpath != nil {
		f, err := os.Create(*outpath)
		panicOn(err)
		_, err = f.Write(by)
		panicOn(err)
		f.Close()
	}

	return
}
