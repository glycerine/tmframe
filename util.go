package tm

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"os"
	"time"
)

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

func GenerateSeriesWithRepeats(reps []int) *Series {

	var frames []*Frame
	t0, err := time.Parse(time.RFC3339, "2016-02-16T00:00:00Z")
	panicOn(err)
	t0 = t0.UTC()

	var f0 *Frame
	for i := range reps {
		j := reps[i]
		jtm := t0.Add(time.Second * time.Duration(i))
		for k := 0; k < j; k++ {
			f0, err = NewFrame(jtm, EvOneFloat64, float64(i), 0, nil)
			panicOn(err)
			frames = append(frames, f0)
		}
	}
	return NewSeriesFromFrames(frames)
}

// each of the nFrame frames is 'spacing' apart, starting at tm
func MakeSequentialTestFrames(nFrame int, tm time.Time,
	spacing time.Duration) []*Frame {
	frames, _, _ := GenTestFrames(nFrame, nil)

	//p("tm = %v", tm)
	for i := range frames {
		t := tm.Add(time.Duration(i) * spacing)
		frames[i].SetTm(TimeToPrimTm(t))
		//p("frame[%v] = %s", i, frames[i])
	}

	return frames
}
